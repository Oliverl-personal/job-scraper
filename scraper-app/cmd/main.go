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
	info    string = "Info: "
	err     string = "Error: "
	debug   string = "Debug: "
	warning string = "Warning: "
)

type JobPosting struct {
	Title          string
	Company        string
	Location       string
	JobDescription string
}

const TITLE_KEY string = "Title"
const COMPANY_KEY string = "Company"
const LOCATION_KEY string = "Location"
const JOB_DESCRIPTION_KEY string = "JobDescription"

// REQUIRES: 	none
// MODIFIES: 	none
// EFFECTS: 	logs error
func LogErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// REQUIRES: 	none
// MODIFIES: 	none
// EFFECTS: 	returns true, nil if map is not empty and balanced, otherwise return false, reason
func CheckOutputJobs(j map[string][]string) (bool, error) {
	if len(j) == 0 {
		return false, errors.New("Input map is empty.")
	}

	return IsMapBalanced(j)
}

// REQUIRES: 	none
// MODIFIES: 	none
// EFFECTS: 	returns true, nil if map is balanced, otherwise return false, reason
func IsMapBalanced(m map[string][]string) (bool, error) {
	lengths := map[string]int{}
	for key, _ := range m {
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
				return isBalanced, errors.New(fmt.Sprintf("Unbalanced Map - Key: %s, Size: %d, Key: %s, Size: %d", prevKey, prevLen, key, length))
			}
		}
	}

	return isBalanced, nil
}

// REQUIRES: 		none
// MODIFIESES: 	strs
// EFFECTS:			return a modified string slice for string matching
func RegexPrep(strs []string) ([]string, error) {
	if len(strs) == 0 {
		return strs, errors.New("Keyword string slice is empty")
	}
	for i, _ := range strs {
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
	keywords, err := RegexPrep(keywords)
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
	keywords, err := RegexPrep(keywords)
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
	c := colly.NewCollector(
		colly.MaxDepth(1), //crawl depth one
	)
	return c
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			returns a jobs postings
func ScrapeJobs(url string, headerK string, headerVal string, selector JobPosting, c *colly.Collector, wg *sync.WaitGroup) map[string][]string {
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set(headerK, headerVal)
		log.Printf("%s Accessing site %s\n", info, url)
	})

	c.OnError(func(r *colly.Response, err error) {
		LogErr(err)
	})

	// scrub site based selector
	jobs := map[string][]string{}
	v := reflect.ValueOf(selector)
	if v.NumField() < 1 {
		log.Fatalf("%s Selector does not have any fields", err)
	}

	for i := 0; i < v.NumField(); i++ {
		// remove closure property of anonymous functions
		j := i
		c.OnHTML(v.Field(j).Interface().(string), func(e *colly.HTMLElement) {
			wg.Add(1)
			jobs[v.Type().Field(j).Name] = append(jobs[v.Type().Field(j).Name], e.Text)
			defer wg.Done()
		})
	}

	c.Visit(url)

	return jobs
}

// REQUIRES: 		none
// MODIFIESES: 	none
// EFFECTS:			convert map[string][]string to []map[string]string
func DataConversion(mapSlice map[string][]string, sliceMap []map[string]string) ([]map[string]string, error) {
	result, err := CheckOutputJobs(mapSlice)
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
	googleSelector := JobPosting{
		"h2.KLsYvd[jsname=\"SBkjJd\"]",
		"div.nJlQNd.sMzDkb",
		"div.sMzDkb:not(.nJlQNd)",
		"span.HBvzbc",
	}

	// googleSelector := JobPosting{
	// 	".whazf h2.KLsYvd",
	// 	".whazf div.nJlQNd.sMzDkb",
	// 	".whazf div.sMzDkb:not(.nJlQNd)",
	// 	".whazf span.HBvzbc",
	// }

	site := "https://www.google.com/search?q=google+jobs&oq=google+jobs&aqs=chrome.0.69i59j0i512j69i59j0i131i433i512j0i512l2j69i60l2.2600j0j7&sourceid=chrome&ie=UTF-8&ibp=htl;jobs&sa=X&ved=2ahUKEwj755zf-53_AhX_GDQIHQ-WBH8Qkd0GegQIDhAB"

	file, err := os.Create(logFile)
	LogErr(err)
	defer file.Close()
	log.SetOutput(file)

	c := InitCollyCollector()

	jobs := ScrapeJobs(site, headerKey, headerValue, googleSelector, c, &wgHTML)
	wgHTML.Wait()

	var unFilteredPostings []map[string]string
	unFilteredPostings, err = DataConversion(jobs, unFilteredPostings)
	LogErr(err)

	// filter twice
	filteredPostings, err = OrFilter(orKeywords, unFilteredPostings)
	LogErr(err)
	filteredPostings, err = AndFilter(andKeywords, filteredPostings)
	LogErr(err)

	// write to file
	log.Printf("%s Adding jobs to %s\n", info, jobsFile)
	f, err := os.Create(jobsFile)
	LogErr(err)
	defer f.Close()

	jobsData, err := json.Marshal(filteredPostings)
	LogErr(err)
	_, err = f.Write(jobsData)
	LogErr(err)

}
