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
	"github.com/influxdata/influxdb/client/v2"
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

	fmt.Println(os.Getenv("MYSQL_DB"))
	fmt.Println(os.Getenv("INFLUX_DB"))
	fmt.Println(os.Getenv("INFLUX_USERNAME"))
	fmt.Println(os.Getenv("INFLUX_PASSWORD"))
	fmt.Println(os.Getenv("INFLUX_ADDRESS"))
	fmt.Println(os.Getenv("PARTY_PAGE"))
	/* Check that we have something in the command line
	This should be the url to scrape
	*/
	if os.Getenv("PARTY_PAGE") == "" {
		log.Fatal("No Url supplied to scrape")
	}
	players = FetchPlayers(os.Getenv("PARTY_PAGE"))
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
	/*
	p_seeking_11 := PlayersBetween( 1,  11, db_players)
	p_seeking_19 := PlayersBetween( 12, 19, db_players)
	p_seeking_29 := PlayersBetween( 20, 29, db_players)
	p_seeking_39 := PlayersBetween( 30, 39, db_players)
	p_seeking_49 := PlayersBetween( 40, 49, db_players)
	p_seeking_59 := PlayersBetween( 50, 59, db_players)
	p_seeking_69 := PlayersBetween( 60, 69, db_players)
	p_seeking_74 := PlayersBetween( 70, 74, db_players)
	p_seeking_75 := PlayersBetween( 75, 76, db_players)
	*/

	group_crying := time.Duration(0)
	for i:= range(db_players) {
		crying_since := db_players[i].Lastseen.Sub(db_players[i].Started_seeking)
		group_crying = group_crying + crying_since
	}

	// If you want to divide a Duration by some variable, you do it like this.
	fmt.Printf("Total crying %s Average time to cry is %s\n", group_crying, group_crying/time.Duration(len(db_players)))

	/*
	influxconfig := make(client.HTTPConfig)
	influxconfig.Addr     = os.Getenv("INFLUX_ADDRESS")
	influxconfig.Username = os.Getenv("INFLUX_USERNAME")
	influxconfig.Password = os.Getenv("INFLUX_PASSWORD")
	influxconfig.Timeout  = time.Duration(60)
	*/
	conn, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://monitoring.imagestream.com:8086",
		Username: "josh",
		Password: "fox",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	q := client.NewQuery("Select last(\"value\") FROM \"nasomi\" WHERE(\"location\" = 'Nasomi' AND \"stat\" = 'population')", "nasomi", "s")
	if response, err := conn.Query(q); err == nil && response.Error() == nil {
		fmt.Printf("Number of items in Result %d\n", len(response.Results))
		fmt.Println(response.Results)
		fmt.Println(response.Results[0].Series[0].Values[0][1])
	} else {
		fmt.Println(err)
		fmt.Println(response.Error())
	}















}
