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
	//"reflect"
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

func PlayersBetween(low_l int, high_l int, db []*Player) int {
	// Find the number of players between a range of levels
	x := 0
	for i := range(db) {
		if (db[i].Mainlevel >= low_l) && (db[i].Mainlevel <= high_l) {
			x++
		}
	}
	return x
}

func main() {
	players    := []*Player{}
	db_players := []*Player{}

	/* Check that we have something in the command line
	This should be the url to scrape
	*/
	if os.Getenv("PARTY_PAGE") == "" {
		log.Fatal("No Url supplied to scrape")
	}
	players = FetchPlayers(os.Getenv("PARTY_PAGE"))

	for i := range(players) {
		//fmt.Printf("%s is seeking ---> Job is %s\n", players[i].Name, players[i].Jobtxt)
		players[i] = Genjobs(players[i])
	}

	// Connect to the MySql database
	db := ConnectMySql()
	defer db.Close()


	// Open influxdb
	conn := ConnectInfluxdb()
	defer conn.Close()

	/* 
	Get the players in the MySql Database, we pass along what we see from the fetch to 
	update the database. By the time we get it here it is the most complete source of information
	*/
	db_players = GetDb(db, players)

	group_crying := time.Duration(0)
	for i:= range(db_players) {
		crying_since := db_players[i].Lastseen.Sub(db_players[i].Started_seeking)
		group_crying = group_crying + crying_since
	}
	average_crying := group_crying/time.Duration(len(db_players))

	// If you want to divide a Duration by some variable, you do it like this.
	fmt.Printf("Total crying %s Average time to cry is %s\n", group_crying, average_crying)

	// Get the total People online
        Total_online := GetNasomiPop(conn)

	PoverS := Total_online / float64(len(db_players))
	fmt.Printf("Total seeking : %d Ratio of pop/seeking: %1f\n", len(db_players), PoverS )

	// Add PoverS to the Batch point
	WriteInflux2Tfl(conn, "Stats", "Ratio_PS", "Ratio", PoverS)

	// Average Seeking Time 
	WriteInflux2Tfl(conn, "Stats", "Seeking_Time", "Average", average_crying.Seconds())

	/*
	Now we generate the info for the level ranges.
	p_seeking_11 := PlayersBetween( 1,  11, db_players)
	p_seeking_19 := PlayersBetween( 12, 19, db_players)
	p_seeking_29 := PlayersBetween( 20, 29, db_players)
	p_seeking_39 := PlayersBetween( 30, 39, db_players)
	p_seeking_49 := PlayersBetween( 40, 49, db_players)
	p_seeking_59 := PlayersBetween( 50, 59, db_players)
	p_seeking_69 := PlayersBetween( 60, 69, db_players)
	p_seeking_74 := PlayersBetween( 70, 74, db_players)
	p_seeking_75 := PlayersBetween( 75, 76, db_players)

	[]Levels

	*/








}
