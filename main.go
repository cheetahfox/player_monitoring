package main

import (
	"fmt"
	"os"
	"net/http"
	"log"
	"github.com/PuerkitoBio/goquery"
)

func FetchPage(url string) {
	/* This funciton is rather specific to the page I am scraping. We take the url 
	and return a 2d slice with players names and jobs that currently seeking 
	party. 
	*/
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatal("Request Status code %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML to parse
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		// We skip the first two TD's because they don't contain normal data
		if i == 1 {
			return;
		}
		if i == 2 {
			return;
		}
		if (i-1)%4 == 0 {
			fmt.Printf("index is %d, name is ----> %s\n", i, s.Text())
		}
		if (i-2)%4 == 0 {
			fmt.Printf("index is %d, job is ----> %s\n", i, s.Text())
		}
		})
}

func main() {
	// Check that we have something in the command line
	if len(os.Args) > 1 {
		FetchPage(os.Args[1])
	}
}
