package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	scraper "scraper-app/apps/scraper"
	"scraper-app/utils"
)

func init() {
	utils.LoadEnv("../.env")
	utils.InitLogger()
}

func main() {

	var jobsDir string = utils.GetEnv("JOBS_DIR")
	// Fields
	andKeywords := []string{"the", "a"}
	orKeywords := []string{"TECH", "SOFT"}

	var filteredPostings []map[string]string

	var wgHTML sync.WaitGroup
	c := scraper.InitCollyCollector()

	googleSelector := scraper.JobPosting{
		Title:          "ul>li:nth-of-type(1) h2.KLsYvd",
		Company:        "ul>li:nth-of-type(1) div.nJlQNd.sMzDkb",
		Location:       "ul>li:nth-of-type(1) div.sMzDkb:not(.nJlQNd)",
		JobDescription: "ul>li:nth-of-type(1) span.HBvzbc",
	}

	googleScraper := scraper.JobScraper{
		Url:         "https://www.google.com/search?q=google+jobs&oq=google+jobs&aqs=chrome.0.69i59j0i512j69i59j0i131i433i512j0i512l2j69i60l2.2600j0j7&sourceid=chrome&ie=UTF-8&ibp=htl;jobs&sa=X&ved=2ahUKEwj755zf-53_AhX_GDQIHQ-WBH8Qkd0GegQIDhAB",
		HeaderKey:   scraper.HeaderKey,
		HeaderValue: scraper.HeaderValue,
		JobSelector: googleSelector,
		CPtr:        c,
		WgPtr:       &wgHTML,
	}

	jobs := scraper.ScrapeJobs(googleScraper)
	wgHTML.Wait()

	var unFilteredPostings []map[string]string
	unFilteredPostings, err := scraper.DataConversion(jobs.JobsInfo, unFilteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}

	// filter twice
	filteredPostings, err = scraper.OrFilter(orKeywords, unFilteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
	filteredPostings, err = scraper.AndFilter(andKeywords, filteredPostings)
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
