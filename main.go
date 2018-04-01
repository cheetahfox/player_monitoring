package main

import (
	"fmt"
	"os"
	"net/http"
	"log"
	"github.com/PuerkitoBio/goquery"
)

func FetchPage() {
	// Get the Url passed into the program
	res, err := http.Get(os.Args[1])
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
	// Check that 
	if len(os.Args) > 1 {
		fmt.Println(os.Args[1])
		FetchPage()
	}
}
