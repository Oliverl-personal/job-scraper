package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"scraper-app/apps/scraper"
	"scraper-app/apps/tagger"
	"scraper-app/utils"
)

func init() {
	utils.LoadEnv("../.env")
}

func main() {
	utils.InitLogger()
	tagger.InitTagger()
	var jobsDir string = utils.GetEnv("JOBS_DIR")
	andKeywords := []string{"the", "a"}
	orKeywords := []string{"TECH", "SOFT"}

	var filteredPostings []map[string]string

	// Scraping
	var wgHTML sync.WaitGroup
	c := scraper.InitCollyCollector()

	googleSelector := scraper.JobPosting{
		Title:          "ul>li:nth-of-type(1) h2.KLsYvd",
		Company:        "ul>li:nth-of-type(1) div.nJlQNd.sMzDkb",
		Location:       "ul>li:nth-of-type(1) div.sMzDkb:not(.nJlQNd)",
		JobDescription: "ul>li:nth-of-type(1) span.HBvzbc",
	}

	googleScraper := scraper.JobScraper{
		Url:         "https://www.google.com/search?q=Software+Engineer+(Kubernetes)&ibp=htl;jobs&sa=X&ved=2ahUKEwjUjoCjqo2AAxWiKX0KHS71CbYQkd0GegQIHxAB#fpstate=tldetail&sxsrf=AB5stBhKdMVijf2ealc389ZerIZtawopnA:1689307884565&htivrt=jobs&htidocid=dUFtP4LbkrIAAAAAAAAAAA%3D%3D",
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

	filteredPostings, err = scraper.OrFilter(orKeywords, unFilteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
	filteredPostings, err = scraper.AndFilter(andKeywords, filteredPostings)
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}

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

	// Tagging
	jobTags, err := tagger.TagJobDescription(filteredPostings[0]["JobDescription"])
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
	fmt.Println(jobTags)

}
