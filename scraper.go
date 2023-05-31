package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/gocolly/colly"
)

const TITLE_KEY string = "title"
const COMPANY_KEY string = "company"
const LOCATION_KEY string = "location"
const JOB_DESCRIPTION_KEY string = "jobDescription"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func isMapBalanced(m map[string][]string) bool {
	var lengths []int
	for key, _ := range m {
		lengths = append(lengths, len(m[key]))
	}

	isBalanced := true
	var tmp int
	for i := range lengths {
		if tmp == 0 {
			tmp = lengths[i]
		} else {
			if tmp == lengths[i] {
				tmp = lengths[i]
			} else {
				isBalanced = false
				break
			}
		}
	}

	return isBalanced
}

func regexPrep(strs []string) []string {
	for i := range strs {
		strs[i] = ".(?i)" + strs[i] + "."
	}
	return strs
}

func andFilter(and []string, unfilteredJobs []map[string]string) ([]map[string]string, error) {
	var filteredJobs []map[string]string
	and = regexPrep(and)

	for i := range unfilteredJobs {
		add := true
		for j := range and {
			contains, e := regexp.MatchString(and[j], unfilteredJobs[i][JOB_DESCRIPTION_KEY])
			if e != nil {
				return filteredJobs, e
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

func orFilter(or []string, unfilteredJobs []map[string]string) ([]map[string]string, error) {
	var filteredJobs []map[string]string
	// prepare for regex
	or = regexPrep(or)

	// filter
	for i := range unfilteredJobs {
		matched, err := regexp.MatchString(or[0], unfilteredJobs[i][JOB_DESCRIPTION_KEY])
		if err != nil {
			return filteredJobs, err
		}

		if matched {
			filteredJobs = append(filteredJobs, unfilteredJobs[i])
		} else {
			for j := 1; len(or) > j; j++ {
				matchedJ, e := regexp.MatchString(or[j], unfilteredJobs[i][JOB_DESCRIPTION_KEY])
				if e != nil {
					return filteredJobs, err
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

func containsKeyword(kw string, str string) bool {
	contains := false

	return contains
}

func main() {
	// Initialize
	andKeywords := []string{"the", "a"}
	orKeywords := []string{"TECH", "SOFT"}
	var unFilteredPostings []map[string]string
	var filteredPostings []map[string]string

	var wgHTML sync.WaitGroup
	googleJobSelector := map[string]string{
		TITLE_KEY:           "h2.KLsYvd[jsname=\"SBkjJd\"]",
		COMPANY_KEY:         "div.nJlQNd.sMzDkb",
		LOCATION_KEY:        "div.sMzDkb:not(.nJlQNd)",
		JOB_DESCRIPTION_KEY: "span.HBvzbc",
	}

	url := "https://www.google.com/search?q=google+jobs&oq=google+jobs&aqs=chrome.0.69i59j0i512j69i59j0i131i433i512j0i512l2j69i60l2.2600j0j7&sourceid=chrome&ie=UTF-8&ibp=htl;jobs&sa=X&ved=2ahUKEwj755zf-53_AhX_GDQIHQ-WBH8Qkd0GegQIDhAB#fpstate=tldetail&htivrt=jobs&htiq=google+jobs&htidocid=t9Q5Jb6QrLcAAAAAAAAAAA%3D%3D&sxsrf=APwXEddxiGmYcOLYx4Ch3xy2ZKw-5YzDAg:1685481463433"
	c := colly.NewCollector(
		colly.MaxDepth(1), //crawl depth one
	)

	// scrub site based selector
	jobs := map[string][]string{}
	for key, selector := range googleJobSelector {
		// remove closure property of anonymous functions
		k := key
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			wgHTML.Add(1)
			jobs[k] = append(jobs[k], e.Text)
			defer wgHTML.Done()
		})
	}

	//print visiting... before making request
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Error while scraping:", e)
	})

	c.Visit(url)
	wgHTML.Wait()

	// organize posting to:
	// "title":
	// "company":
	// "location":
	// "jobDescription":

	isBalanced := isMapBalanced(jobs)
	if !isBalanced {
		panic("Jobs is not balanced")
	}

	length := len(jobs[TITLE_KEY])

	for i := 0; length > i; i++ {
		jobPosting := make(map[string]string)
		for key, value := range jobs {
			jobPosting[key] = value[i]
		}
		unFilteredPostings = append(unFilteredPostings, jobPosting)
	}

	var e error

	// filter twice
	filteredPostings, e = orFilter(orKeywords, unFilteredPostings)
	check(e)
	filteredPostings, e = andFilter(andKeywords, filteredPostings)
	check(e)

	// write to file
	f, e := os.Create("tmp/jobs.json")
	check(e)
	defer f.Close()

	jobsData, e := json.Marshal(filteredPostings)
	check(e)
	_, err1 := f.Write(jobsData)
	check(err1)

}
