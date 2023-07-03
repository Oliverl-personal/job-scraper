package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sync"

	"scraper-app/utils"

	"github.com/gocolly/colly"
)

const (
	headerKey   string = "User-Agent"
	headerValue string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
)

type jobPosting struct {
	Title          string
	Company        string
	Location       string
	JobDescription string
}

// (site, headerKey, headerValue, googleSelector, c, &wgHTML)

type jobScraper struct {
	url         string
	headerKey   string
	headerValue string
	jobSelector jobPosting
	cPtr        *colly.Collector
	wgPtr       *sync.WaitGroup
}

const (
	TITLE_KEY           string = "Title"
	COMPANY_KEY         string = "Company"
	LOCATION_KEY        string = "Location"
	JOB_DESCRIPTION_KEY string = "JobDescription"
)

type jobs struct {
	mu       sync.Mutex
	jobsInfo map[string][]string
}

// REQUIRES: 	none
// MODIFIES: 	none
// EFFECTS: 	returns true, nil if map is not empty and balanced, otherwise return false, reason
func checkScrapedJobs(j map[string][]string) (bool, error) {
	if len(j) == 0 {
		return false, errors.New("scraped jobs information is empty")
	}

	return checkJobInfoCompleteness(j)
}

// Job information are split up into categories: job title, job description, job location... etc
// checks if len(job_title) == len(job_description) == len(job_location) ...
func checkJobInfoCompleteness(m map[string][]string) (bool, error) {
	lengths := map[string]int{}
	for key := range m {
		lengths[key] = len(m[key])
	}

	isBalanced := true
	var prevLen int
	var prevKey string
	for key, length := range lengths {
		if prevLen == 0 {
			prevLen = length
			prevKey = key
		} else {
			if prevLen == lengths[key] {
				prevLen = lengths[key]
				prevKey = key

			} else {
				isBalanced = false
				return isBalanced, fmt.Errorf("scraped job information incomplete: key: %s, size: %d, key: %s, size: %d", prevKey, prevLen, key, length)
			}
		}
	}

	return isBalanced, nil
}

// REQUIRES: 		none
// MODIFIESES: 	strs
// EFFECTS:			return a modified string slice for string matching
func regexPrep(strs []string) ([]string, error) {
	if len(strs) == 0 {
		return strs, errors.New("keyword string slice is empty")
	}
	for i := range strs {
		// match any ".", not case sensitive (?i)
		strs[i] = "(?i)" + strs[i]
	}
	return strs, nil
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			error or filteredJobs that includes all (add) keywords provided based on the job description
func andFilter(keywords []string, unfilteredJobs []map[string]string) ([]map[string]string, error) {
	var filteredJobs []map[string]string
	keywords, err := regexPrep(keywords)
	if err != nil {
		return nil, err
	}
	for i := range unfilteredJobs {
		add := true
		for j := range keywords {
			contains, err := regexp.MatchString(keywords[j], unfilteredJobs[i][JOB_DESCRIPTION_KEY])
			if err != nil {
				return nil, err
			}
			if !contains {
				add = false
				break
			}
		}
		if add {
			filteredJobs = append(filteredJobs, unfilteredJobs[i])
		}
	}

	return filteredJobs, nil
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			error or filteredJobs that includes any (or) keywords provided based on the job description
func orFilter(keywords []string, unfilteredJobs []map[string]string) ([]map[string]string, error) {
	var filteredJobs []map[string]string
	// prepare keyword for regex string match
	keywords, err := regexPrep(keywords)
	if err != nil {
		return nil, err
	}

	// filter
	for i := range unfilteredJobs {
		matched, err := regexp.MatchString(keywords[0], unfilteredJobs[i][JOB_DESCRIPTION_KEY])
		if err != nil {
			return nil, err
		}

		if matched {
			filteredJobs = append(filteredJobs, unfilteredJobs[i])
		} else {
			for j := 1; len(keywords) > j; j++ {
				matchedJ, e := regexp.MatchString(keywords[j], unfilteredJobs[i][JOB_DESCRIPTION_KEY])
				if e != nil {
					return nil, err
				}
				if matchedJ {
					filteredJobs = append(filteredJobs, unfilteredJobs[i])
					break
				}
			}
		}
	}

	return filteredJobs, nil
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			Initializes colly collector
func initCollyCollector() *colly.Collector {
	c := colly.NewCollector(colly.MaxDepth(1))
	return c
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			returns a jobs postings
func scrapeJobs(js jobScraper) jobs {
	js.cPtr.OnRequest(func(r *colly.Request) {
		r.Headers.Set(js.headerKey, js.headerValue)
		utils.Logger.Debug(fmt.Sprintf("accessing site %s", js.url))
	})

	js.cPtr.OnError(func(r *colly.Response, err error) {
		utils.FatalError(fmt.Errorf("%v", err))
	})

	// scrub site based selector
	var jobs jobs
	jobs.jobsInfo = make(map[string][]string)
	v := reflect.ValueOf(js.jobSelector)
	if v.NumField() < 1 {
		utils.FatalError(fmt.Errorf("selector does not have any fields"))
	}

	for i := 0; i < v.NumField(); i++ {
		// remove closure property of anonymous functions
		j := i
		js.cPtr.OnHTML(v.Field(j).Interface().(string), func(e *colly.HTMLElement) {
			js.wgPtr.Add(1)
			jobs.mu.Lock()
			defer jobs.mu.Unlock()
			jobs.jobsInfo[v.Type().Field(j).Name] = append(jobs.jobsInfo[v.Type().Field(j).Name], e.Text)
			defer js.wgPtr.Done()
		})
	}

	err := js.cPtr.Visit(js.url)
	if err != nil {
		utils.FatalError(err)
	}

	return jobs
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			convert map[string][]string to []map[string]string
func dataConversion(mapSlice map[string][]string, sliceMap []map[string]string) ([]map[string]string, error) {
	result, err := checkScrapedJobs(mapSlice)
	if !result {
		return nil, err
	}

	length := len(mapSlice[TITLE_KEY])
	for i := 0; length > i; i++ {
		jobPosting := make(map[string]string)
		for key, value := range mapSlice {
			jobPosting[key] = value[i]
		}
		sliceMap = append(sliceMap, jobPosting)
	}

	return sliceMap, nil
}

func main() {
	utils.LoadEnv("../.env")
	var jobsDir string = utils.GetEnv("JOBS_DIR")
	// Fields
	andKeywords := []string{"the", "a"}
	orKeywords := []string{"TECH", "SOFT"}

	var filteredPostings []map[string]string

	var wgHTML sync.WaitGroup
	c := initCollyCollector()

	googleSelector := jobPosting{
		"ul>li:nth-of-type(1) h2.KLsYvd",
		"ul>li:nth-of-type(1) div.nJlQNd.sMzDkb",
		"ul>li:nth-of-type(1) div.sMzDkb:not(.nJlQNd)",
		"ul>li:nth-of-type(1) span.HBvzbc",
	}

	googleScraper := jobScraper{
		"https://www.google.com/search?q=google+jobs&oq=google+jobs&aqs=chrome.0.69i59j0i512j69i59j0i131i433i512j0i512l2j69i60l2.2600j0j7&sourceid=chrome&ie=UTF-8&ibp=htl;jobs&sa=X&ved=2ahUKEwj755zf-53_AhX_GDQIHQ-WBH8Qkd0GegQIDhAB",
		headerKey,
		headerValue,
		googleSelector,
		c,
		&wgHTML,
	}

	jobs := scrapeJobs(googleScraper)
	wgHTML.Wait()

	var unFilteredPostings []map[string]string
	unFilteredPostings, err := dataConversion(jobs.jobsInfo, unFilteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}

	// filter twice
	filteredPostings, err = orFilter(orKeywords, unFilteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
	filteredPostings, err = andFilter(andKeywords, filteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}

	// write to file
	utils.Logger.Debug(fmt.Sprintf("adding jobs to jobs file"))
	_, err = os.Stat(jobsDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(jobsDir, 0777)
		if err != nil {
			utils.FatalError(fmt.Errorf("unable to create jobs directory"))
		}
	}
	jobsFile := jobsDir + "/jobs.json"
	f, err := os.Create(jobsFile)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
	defer func() {
		err := f.Close()
		if err != nil {
			utils.FatalError(fmt.Errorf("%v, ", err))
		}
	}()

	jobsData, err := json.Marshal(filteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
	_, err = f.Write(jobsData)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
}
