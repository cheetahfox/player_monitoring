package main

import (
	"fmt"
	"os"
	"net/http"
	"log"
	"strings"
	"regexp"
	"strconv"
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/PuerkitoBio/goquery"
)

// This is the main user data structure
type Player struct {
	Name string
	Jobtxt string
	Mainjob string
	Mainlevel int
	Subjob string
	Sublevel int
	Lastseen time.Time
	Started_seeking time.Time
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

	now := time.Now()
	for x := 0; x < len(names); x++ {
		newplayer := new(Player)
		newplayer.Name = names[x]
		newplayer.Jobtxt = jobs[x]
		newplayer.Lastseen = now
		players = append(players, newplayer)
	}

	return players
}

func Genjobs(user *Player) *Player {
	// Split the jobtxt
	jobs := strings.Split(user.Jobtxt, "/")

	/* 
	Since our data always has a / in it, we know we will have a slice with two strings
	Index[0] is always going to be the mainjob level and [1] will always be the subjob
	also setup the regex to match only numbers
	*/
	re := regexp.MustCompile("[0-9]+")
	if mlevel, err := strconv.Atoi(re.FindString(jobs[0])); err == nil {
		user.Mainlevel = mlevel
	}
	if slevel, err := strconv.Atoi(re.FindString(jobs[1])); err == nil {
		user.Sublevel = slevel
	}

	re = regexp.MustCompile("[a-zA-Z]+")
	user.Mainjob = re.FindString(jobs[0])
	user.Subjob =  re.FindString(jobs[1])
	return user
}

func PlayerinDB( player *Player, db []*Player) (bool) {
	/* 
	This function checks to see if the player is in a different 
	*/
	for index := range(db) {
		if player.Name == db[index].Name {
			if player.Mainjob == db[index].Mainjob {
				return true
			}
		}
	}
	return false
}

func main() {
	players    := []*Player{}
	db_players := []*Player{}

	fmt.Println(os.Getenv("MYSQL_DB"))
	fmt.Println(os.Getenv("INFLUX_DB"))
	fmt.Println(os.Getenv("PARTY_PAGE"))
	/* Check that we have something in the command line
	This should be the url to scrape
	*/
	if os.Getenv("PARTY_PAGE") != "" {
		players = FetchPlayers(os.Getenv("PARTY_PAGE"))
	}
	fmt.Printf("%d Players seeking\n",len(players))

	for i := range(players) {
		//fmt.Printf("%s is seeking ---> Job is %s\n", players[i].Name, players[i].Jobtxt)
		players[i] = Genjobs(players[i])
	}
	/*
	Check if we have the MySql enviromental variable set, Open the database if we do
	*/
	if os.Getenv("MYSQL_DB") == "" {
		log.Fatal("No Mysql DB definded")
	}

        db, err := sql.Open("mysql", os.Getenv("MYSQL_DB"))
        if err != nil {
                log.Fatal("MySql DB failure %s", err)
	        panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	/* 
	Get the players in the MySql Database, we pass along what we see from the fetch to 
	update the database. By the time we get it here it is the most complete source of information
	*/
	db_players = GetDb(db, players)

	fmt.Println("about to loop")
	for i:= range(db_players) {
		crying_since := db_players[i].Started_seeking.Sub(db_players[i].Lastseen)
		fmt.Printf("%s is seeking and they have been seeking %s\n", db_players[i].Name, crying_since)
	}






}
