package main

import (
	"fmt"
	"os"
	"net/http"
	"log"
	"database/sql"
	"strings"
	"regexp"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/PuerkitoBio/goquery"
)

// This is the main user data structure
type Player struct {
	Name string
	Jobtxt string
	mainjob string
	mainlevel int
	subjob string
	sublevel int
}

func FetchPlayers(url string) []*Player {
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

	players := []*Player{}
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
		/* This is going to give us two slices where the index matches up each players
		jobs and names are the same. Since we aren't modifying these in later in the program 
		I think this is ok...
		*/
		if (i-1)%4 == 0 {
			names = append(names, s.Text())
		}
		if (i-2)%4 == 0 {
			jobs = append(jobs, s.Text())
		}
		})

	for x := 0; x < len(names); x++ {
		newplayer := new(Player)
		newplayer.Name = names[x]
		newplayer.Jobtxt = jobs[x]
		players = append(players, newplayer)
	}

	return players
}

func Genjobs(user *Player) *Player {
	// Split the jobtxt and 
	jobs := strings.Split(user.Jobtxt, "/")

	/* 
	Since our data always has a / in it, we know we will have a slice with two strings
	Index[0] is always going to be the mainjob level and [1] will always be the subjob
	also setup the regex to match only numbers
	*/
	re := regexp.MustCompile("[0-9]+")
	if mlevel, err := strconv.Atoi(re.FindString(jobs[0])); err == nil {
		user.mainlevel = mlevel
	}
	if slevel, err := strconv.Atoi(re.FindString(jobs[1])); err == nil {
		user.sublevel = slevel
	}

	re = regexp.MustCompile("[a-zA-Z]+")
	user.mainjob = re.FindString(jobs[0])
	user.subjob =  re.FindString(jobs[1])
	return user
}

func GetDB(conn string) {
	/*
	This functtion is connects to the mysql database to store player names peristently 
	The connection information is passed in as a commandline option(for now) format is...
	<username>:<pw>@tcp(<HOST>:<port>)/<dbname>
	This will be passed in through k8s in the end.
	*/
	db, err := sql.Open("mysql", conn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Database Open\n)")
}

func main() {
	players := []*Player{}

	/* Check that we have something in the command line
	This should be the url to scrape
	*/
	if len(os.Args) > 1 {
		players = FetchPlayers(os.Args[1])
	}
	fmt.Printf("%d Players seeking\n",len(players))

	for i := range(players) {
		//fmt.Printf("%s is seeking ---> Job is %s\n", players[i].Name, players[i].Jobtxt)
		players[i] = Genjobs(players[i])
	}
	if len(os.Args) > 2 {
		GetDB(os.Args[2])
	}

}
