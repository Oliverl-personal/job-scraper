package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sync"

	"github.com/gocolly/colly"
)

const (
	logFile     string = "../tmp/log.txt"
	jobsFile    string = "../tmp/jobs.json"
	headerKey   string = "User-Agent"
	headerValue string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
)

const (
	info string = "Info: "
	err  string = "Error: "
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
// EFFECTS: 	logs error
func logErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// REQUIRES: 	none
// MODIFIES: 	none
// EFFECTS: 	returns true, nil if map is not empty and balanced, otherwise return false, reason
func checkOutputJobs(j map[string][]string) (bool, error) {
	if len(j) == 0 {
		return false, errors.New("input map is empty")
	}

	return isMapBalanced(j)
}

// REQUIRES: 	none
// MODIFIES: 	none
// EFFECTS: 	returns true, nil if map is balanced, otherwise return false, reason
func isMapBalanced(m map[string][]string) (bool, error) {
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
				return isBalanced, fmt.Errorf("Unbalanced Map - Key: %s, Size: %d, Key: %s, Size: %d", prevKey, prevLen, key, length)
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
		return strs, errors.New("Keyword string slice is empty")
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
		log.Printf("%s Accessing site %s\n", info, js.url)
	})

	js.cPtr.OnError(func(r *colly.Response, err error) {
		logErr(err)
	})

	// scrub site based selector
	var jobs jobs
	jobs.jobsInfo = make(map[string][]string)
	v := reflect.ValueOf(js.jobSelector)
	if v.NumField() < 1 {
		log.Fatalf("%s Selector does not have any fields", err)
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
	logErr(err)

	return jobs
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			convert map[string][]string to []map[string]string
func dataConversion(mapSlice map[string][]string, sliceMap []map[string]string) ([]map[string]string, error) {
	result, err := checkOutputJobs(mapSlice)
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
	// Fields
	andKeywords := []string{"the", "a"}
	orKeywords := []string{"TECH", "SOFT"}

	var filteredPostings []map[string]string

	var wgHTML sync.WaitGroup
	c := initCollyCollector()

	googleSelector := jobPosting{
		"h2.KLsYvd[jsname=\"SBkjJd\"]",
		"div.nJlQNd.sMzDkb",
		"div.sMzDkb:not(.nJlQNd)",
		"span.HBvzbc",
	}

	googleScraper := jobScraper{
		"https://www.google.com/search?q=google+jobs&oq=google+jobs&aqs=chrome.0.69i59j0i512j69i59j0i131i433i512j0i512l2j69i60l2.2600j0j7&sourceid=chrome&ie=UTF-8&ibp=htl;jobs&sa=X&ved=2ahUKEwj755zf-53_AhX_GDQIHQ-WBH8Qkd0GegQIDhAB",
		headerKey,
		headerValue,
		googleSelector,
		c,
		&wgHTML,
	}

	// googleSelector := JobPosting{
	// 	".whazf h2.KLsYvd",
	// 	".whazf div.nJlQNd.sMzDkb",
	// 	".whazf div.sMzDkb:not(.nJlQNd)",
	// 	".whazf span.HBvzbc",
	// }

	file, err := os.Create(logFile)
	logErr(err)
	defer func() {
		err := file.Close()
		logErr(err)
	}()
	log.SetOutput(file)

	jobs := scrapeJobs(googleScraper)
	wgHTML.Wait()

	var unFilteredPostings []map[string]string
	unFilteredPostings, err = dataConversion(jobs.jobsInfo, unFilteredPostings)
	logErr(err)

	// filter twice
	filteredPostings, err = orFilter(orKeywords, unFilteredPostings)
	logErr(err)
	filteredPostings, err = andFilter(andKeywords, filteredPostings)
	logErr(err)

	// write to file
	log.Printf("%s Adding jobs to %s\n", info, jobsFile)
	f, err := os.Create(jobsFile)
	logErr(err)
	defer func() {
		err := f.Close()
		logErr(err)
	}()

	jobsData, err := json.Marshal(filteredPostings)
	logErr(err)
	_, err = f.Write(jobsData)
	logErr(err)
}
