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

type P_Dist struct {
	Player_Level int
	pop string
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

func FetchDistribution(url string) []*P_Dist {
	/*
	This function returns a slice of Player Distributions
	*/
	Distribution := []*P_Dist{}

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
	var Level int
	var Pop string
	switch i {
		// Levels < 12
		case 27:
			Level = 11
			Pop   = s.Text()
		// Levels 12 - 19
		case 28:
			Level = 19
			Pop   = s.Text()
		// Levels 20 - 29 
		case 29:
			Level = 29
			Pop   = s.Text()
		// Levels 30 - 39
		case 30:
			Level = 39
			Pop   = s.Text()
		// Levels 40 - 49
		case 31:
			Level = 49
			Pop   = s.Text()
		// Levels 50 - 59
		case 32:
			Level = 59
			Pop   = s.Text()
		// Levels 60 - 69
		case 33:
			Level = 69
			Pop   = s.Text()
		// Levels 70 - 74
		case 34:
			Level = 74
			Pop   = s.Text()
		// Level 75
		case 35:
			Level = 75
			Pop   = s.Text()
	}
	if Level != 0 {
		newdist := new(P_Dist)
		newdist.Player_Level = Level
		newdist.pop   = Pop
		Distribution = append(Distribution, newdist)
	}
	})
	return Distribution
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

func PlayersBetween(low_l int, high_l int, db_players []*Player) int {
	// Find the number of players between a range of levels
	x := 0
	for i := range(db_players) {
		if (db_players[i].Mainlevel >= low_l) && (db_players[i].Mainlevel <= high_l) {
			x++
		}
	}
	return x
}

func SeekingDistribution(db_players []*Player) []*P_Dist {
	Distribution := []*P_Dist{}
	Level_Ranges := []int{11,19,29,39,49,59,69,74,75}
	/*
	Now we generate the info for the level ranges.
	*/
	for i:= range(Level_Ranges) {
		/* 
		We are going to do a few different things based on which range we are looking.
		Most of the levels are handled by the default case but, 1-11, 12-19, 70-47 and 75
		need different logic. I don't like this but I think this is the cleanest way to handle
		this code. Out of the nine possible cases five of them are handled by the default case
		*/
		switch Level_Ranges[i] {
			case 11:
				newdist := new(P_Dist)
				newdist.Player_Level = Level_Ranges[i]
				newdist.pop = strconv.Itoa(PlayersBetween( 1,  Level_Ranges[i], db_players))
				Distribution = append(Distribution, newdist)
			case 19:
				newdist := new(P_Dist)
				newdist.Player_Level = Level_Ranges[i]
				newdist.pop = strconv.Itoa(PlayersBetween( 12,  Level_Ranges[i], db_players))
				Distribution = append(Distribution, newdist)
			case 74:
				newdist := new(P_Dist)
				newdist.Player_Level = Level_Ranges[i]
				newdist.pop = strconv.Itoa(PlayersBetween( 70,  Level_Ranges[i], db_players))
				Distribution = append(Distribution, newdist)
			case 75:
				newdist := new(P_Dist)
				newdist.Player_Level = Level_Ranges[i]
				newdist.pop = strconv.Itoa(PlayersBetween( Level_Ranges[i], 76 , db_players))
				Distribution = append(Distribution, newdist)
			default:
				newdist := new(P_Dist)
				newdist.Player_Level = Level_Ranges[i]
				newdist.pop = strconv.Itoa(PlayersBetween( Level_Ranges[i]-9,  Level_Ranges[i], db_players))
				Distribution = append(Distribution, newdist)
		}
	}
	return Distribution
}

func main() {
	players    := []*Player{}
	db_players := []*Player{}
	player_distribution := []*P_Dist{}
	seeking_distribution := []*P_Dist{}
	fmt.Println("Playermon startup version 0.01")

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

	if os.Getenv("STATUS_PAGE") == "" {
		log.Fatal("No Url for status page")
	}
	player_distribution = FetchDistribution(os.Getenv("STATUS_PAGE"))

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

	// Get the player/level distribution of seeking players.
	seeking_distribution = SeekingDistribution(db_players)

	/* 
	Write both of the Player and Seeking Distributions into the influxdb database
	*/
	for i:= range(player_distribution) {
		pd_value, err := strconv.Atoi(player_distribution[i].pop)
		if err == nil {
			player_dist_level := strconv.Itoa(player_distribution[i].Player_Level)
			WriteInflux2Tint(conn, "Stats", "Player_Distribution","Online", "Level", player_dist_level, pd_value)
		} else {
			fmt.Println("Error strcon", err)
		}
	}
	for i:= range(seeking_distribution) {
		sd_value, err := strconv.Atoi(seeking_distribution[i].pop)
		if err == nil {
			seeking_dist_level := strconv.Itoa(seeking_distribution[i].Player_Level)
			WriteInflux2Tint(conn, "Stats", "Player_Distribution", "Seeking", "Level", seeking_dist_level, sd_value)
		}
	}

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
	WriteInflux1Tfl(conn, "Stats", "Ratio_PS", "Ratio", PoverS)

	// Average Seeking Time 
	WriteInflux1Tfl(conn, "Stats", "Seeking_Time", "Average", average_crying.Seconds())



}
