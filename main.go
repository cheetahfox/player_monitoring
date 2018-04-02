package main

import (
	"fmt"
	"os"
	"net/http"
	"log"
	"github.com/PuerkitoBio/goquery"
)

func FetchPage(url string) ([]string, []string) {
	/* This funciton is rather specific to the page I am scraping. We take the url 
	and return two slices with players names and jobs that currently seeking 
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

	names := make([]string,0)
	jobs  := make([]string,0)

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		// We skip the first two TD's because they don't contain normal data
		if i == 1 {
			return;
		}
		if i == 2 {
			return;
		}
		if (i-1)%4 == 0 {
			// x is the index in the data slice
			x := (i-1)/4
			fmt.Printf("index is %d, Slice Element is %d, name is ----> %s\n", i, x, s.Text())
			names = append(names, s.Text())
		}
		if (i-2)%4 == 0 {
			x := (i-2)/4
			fmt.Printf("index is %d, Slice Element is %d, job is ----> %s\n", i, x, s.Text())
			jobs = append(jobs, s.Text())
		}
		})
	return names, jobs
}

func main() {
	names := make([]string,0)
	jobs  := make([]string,0)

	// Check that we have something in the command line
	if len(os.Args) > 1 {
		names, jobs = FetchPage(os.Args[1])
	}
	fmt.Printf("size of names %d\n",len(names))
	fmt.Printf("size of jobs  %d\n",len(jobs))
	/*for _, x := range names {
		fmt.Printf(" Player: %s, is seeking on %s\n", names[x], jobs[x])
	} */
}
