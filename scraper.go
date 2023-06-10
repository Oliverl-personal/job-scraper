package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func LogErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// EFFECTS: Check if http response status
// RETURNS: nil if response == OK, nil otherwise
func CheckRespStatus(r *http.Response) error {
	if r.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Failed to fetch http.Response %d %s", r.StatusCode, r.Status))
	}
	return nil
}

func main() {
	logFile := "tmp/log.txt"
	url := "https://www.google.com/search?q=frontend&oq=google+jobs&aqs=chrome.0.69i59j0i512j69i59j0i131i433i512j0i512l2j69i60l2.2600j0j7&sourceid=chrome&ie=UTF-8&ibp=htl;jobs&sa=X&ved=2ahUKEwj755zf-53_AhX_GDQIHQ-WBH8Qkd0GegQIDhAB#fpstate=tldetail&sxsrf=APwXEddxiGmYcOLYx4Ch3xy2ZKw-5YzDAg:1685481463433&htivrt=jobs&htidocid=ETS2Y0qxZOcAAAAAAAAAAA%3D%3D"

	file, err := os.Create(logFile)
	LogErr(err)
	defer file.Close()
	log.SetOutput(file)

	log.Printf("Accessing site %s\n", url)
	resp, err := http.Get(url)
	LogErr(err)
	defer resp.Body.Close()

	log.Println("Checking site response.")
	LogErr(CheckRespStatus(resp))

	log.Println("Creating Document from response.")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	LogErr(err)

	selector := "#gb_Ce"
	log.Printf("Searching for Selector: %s \n", selector)
	title := doc.Find(selector).Text()
	fmt.Println(title)
}
