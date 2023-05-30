package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

func main() {

	url := "https://www.google.com/search?q=google+jobs&oq=google+jobs&aqs=chrome.0.69i59j0i512j69i59j0i131i433i512j0i512l2j69i60l2.2600j0j7&sourceid=chrome&ie=UTF-8&ibp=htl;jobs&sa=X&ved=2ahUKEwj755zf-53_AhX_GDQIHQ-WBH8Qkd0GegQIDhAB#fpstate=tldetail&htivrt=jobs&htiq=google+jobs&htidocid=t9Q5Jb6QrLcAAAAAAAAAAA%3D%3D&sxsrf=APwXEddxiGmYcOLYx4Ch3xy2ZKw-5YzDAg:1685481463433"

	c := colly.NewCollector(
		colly.MaxDepth(1), //crawl depth one
	)

	// Take every element
	c.OnHTML("*", func(e *colly.HTMLElement) {
		
		fmt.Printf("Text found: %s\n", e.Text)
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Error while scraping:", e)
	})
	
	c.Visit(url)
}
