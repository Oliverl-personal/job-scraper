package app

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"scraper-app/utils"
	"sync"

	"github.com/gocolly/colly"
)

const (
	HeaderKey   string = "User-Agent"
	HeaderValue string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
)

type JobPosting struct {
	Title          string
	Company        string
	Location       string
	JobDescription string
}

// (site, headerKey, headerValue, googleSelector, c, &wgHTML)

type JobScraper struct {
	Url         string
	HeaderKey   string
	HeaderValue string
	JobSelector JobPosting
	CPtr        *colly.Collector
	WgPtr       *sync.WaitGroup
}

const (
	TITLE_KEY           string = "Title"
	COMPANY_KEY         string = "Company"
	LOCATION_KEY        string = "Location"
	JOB_DESCRIPTION_KEY string = "JobDescription"
)

type Jobs struct {
	mu       sync.Mutex
	JobsInfo map[string][]string
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
func AndFilter(keywords []string, unfilteredJobs []map[string]string) ([]map[string]string, error) {
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
func OrFilter(keywords []string, unfilteredJobs []map[string]string) ([]map[string]string, error) {
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
func InitCollyCollector() *colly.Collector {
	c := colly.NewCollector(colly.MaxDepth(1))
	return c
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			returns a jobs postings
func ScrapeJobs(js JobScraper) Jobs {
	js.CPtr.OnRequest(func(r *colly.Request) {
		r.Headers.Set(js.HeaderKey, js.HeaderValue)
		utils.Logger.Debug(fmt.Sprintf("accessing site %s", js.Url))
	})

	js.CPtr.OnError(func(r *colly.Response, err error) {
		utils.FatalError(fmt.Errorf("%v", err))
	})

	// scrub site based selector
	var jobs Jobs
	jobs.JobsInfo = make(map[string][]string)
	v := reflect.ValueOf(js.JobSelector)
	if v.NumField() < 1 {
		utils.FatalError(fmt.Errorf("selector does not have any fields"))
	}

	for i := 0; i < v.NumField(); i++ {
		// remove closure property of anonymous functions
		j := i
		js.CPtr.OnHTML(v.Field(j).Interface().(string), func(e *colly.HTMLElement) {
			js.WgPtr.Add(1)
			jobs.mu.Lock()
			defer jobs.mu.Unlock()
			jobs.JobsInfo[v.Type().Field(j).Name] = append(jobs.JobsInfo[v.Type().Field(j).Name], e.Text)
			defer js.WgPtr.Done()
		})
	}

	err := js.CPtr.Visit(js.Url)
	if err != nil {
		utils.FatalError(err)
	}

	return jobs
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			convert map[string][]string to []map[string]string
func DataConversion(mapSlice map[string][]string, sliceMap []map[string]string) ([]map[string]string, error) {
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
