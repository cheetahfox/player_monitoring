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

const Version = "0.07"

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

type Tod struct {
	NM string
	Killer string
	LinkShell string
	Last_Seen time.Time
	First_Seen time.Time
}

func FetchTods(url string) []*Tod {
	/*
	Grab the ToDs
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

	tods := []*Tod{}
	nms        := make([]string,0)
	killers    := make([]string,0)
	linkshells := make([]string,0)

        doc.Find("td").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return;
		}
                if i == 1 {
                        return;
                }
                if i == 2 {
                        return;
                }
		if i == 3 {
			return;
		}
		if i == 4 {
			return;
		}
		if i == 5 {
			return;
		}
		if i == 6 {
			return;
		}
		if (i-3)%3 == 0 {
			linkshells = append(linkshells, s.Text())
		}
                if (i-4)%3 == 0 {
			nms = append(nms, s.Text())
                }
                if (i-5)%3 == 0 {
			killers = append(killers, s.Text())
                }
                })

	now := time.Now()
	for x := 0; x < len(nms); x++ {
		newtod           := new(Tod)
		newtod.NM         = nms[x]
		newtod.Killer     = killers[x]
		newtod.LinkShell  = linkshells[x]
		newtod.Last_Seen  = now
		tods = append(tods, newtod)
	}

	return tods
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

	// Split the jobtxt into actual jobs
	for i := range(players) {
		//fmt.Printf("%s is seeking ---> Job is %s\n", players[i].Name, players[i].Jobtxt)
		players[i] = Genjobs(players[i])
	}

	return players
}

func  GenerateDistribution(Stats map[string]int) []*P_Dist {
	/*
	This function returns a slice of Player Distributions
	It's a Wraper function that allows the rest of the program to work given that I should have been using 
	maps to store my data when I was parsing the Url (done now in FetchStats). But because I didn't really
	know about Maps (I am new at GO) I was doing it in a more complicated way. Eventially I will rewrite 
	everything else to use maps but for now this takes the map in and generates the Slice of P_Dist's 
	*/

	Distribution := []*P_Dist{}
	Level_Ranges := []int{11,19,29,39,49,59,69,74,75}

	for i:= range(Level_Ranges) {
		newdist := new(P_Dist)
		newdist.Player_Level = Level_Ranges[i]
		Player_Level :=  strconv.Itoa(Level_Ranges[i])
		pop := strconv.Itoa(Stats["Dist_level_" + Player_Level])
		newdist.pop = pop
		Distribution = append(Distribution, newdist)
	}

	return Distribution
}

func FetchStats(url string) map[string]int {
	/*
	This function parses the Nasomi Stats Page. The specific format is driven by Nasomi's page
	If he changes the format, this will need to be adjusted 
	*/
	Stats := make(map[string]int)
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

	// Regex to match non-nummeric chars
        re := regexp.MustCompile("[,a-zA-Z]")

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		switch i {
			// Number of AH Transacatons
			case 6:
				if sales, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["AH_Transactions"] = sales
				}
			// Amount of AH gil exchanged
			case 8:
				if gil, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["AH_gil"] = gil
				}
			// Mobs killed
			case 11:
				if mob_death, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Mob_Deaths"] = mob_death
				}
			// Player Deaths
			case 15:
				if player_death, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Player_Deaths"] = player_death
				}
			// Levels < 12
			case 27:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_11"] = pop
				}
			// Levels 12 - 19
			case 28:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_19"] = pop
				}
			// Levels 20 - 29 
			case 29:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_29"] = pop
				}
			// Levels 30 - 39
			case 30:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_39"] = pop
				}
			// Levels 40 - 49
			case 31:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_49"] = pop
				}
			// Levels 50 - 59
			case 32:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_59"] = pop
				}
			// Levels 60 - 69
			case 33:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_69"] = pop
				}
			// Levels 70 - 74
			case 34:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_74"] = pop
				}
			// Level 75
			case 35:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Dist_level_75"] = pop
				}
			// Total Current population
			case 36:
				if pop, err := strconv.Atoi(re.ReplaceAllString(s.Text(), "")); err == nil {
					Stats["Current_Population"] = pop
				}
		}
		})
	return Stats
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
	fmt.Printf("Playermon startup version %s\n", Version)

	/* Check that we have something in the command line
	This should be the url to scrape
	*/
	if os.Getenv("PARTY_PAGE") == "" {
		log.Fatal("No Url supplied to scrape")
	}
	players = FetchPlayers(os.Getenv("PARTY_PAGE"))

	if os.Getenv("STATUS_PAGE") == "" {
		log.Fatal("No Url for status page")
	}

	// Get stats
	Stats := make(map[string]int)
	Stats = FetchStats(os.Getenv("STATUS_PAGE"))

	// Get ToD's
	Tods := FetchTods(os.Getenv("TOD_PAGE"))

	// Generate player distribution
	player_distribution = GenerateDistribution(Stats)

	// Connect to the MySql database
	db := ConnectMySql()
	defer db.Close()

	// Open influxdb
	conn := ConnectInfluxdb()
	defer conn.Close()

	/*
	Refresh the Tods Database
	After I update it, I don't do anything else with it in this program. It's displayed in Grafana
	*/
	GenTodDb(db, Tods)

	/* 
	Get the players in the MySql Database, we pass along what we see from the fetch to 
	update the database. By the time we get it here it is the most complete source of information
	*/

	db_players = GetDb(db, players)

	// Get the player/level distribution of seeking players.
	seeking_distribution = SeekingDistribution(db_players)

	// Populate the number of people seeing in the Stats map
	Stats["Seeking_Population"] = len(db_players)

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
		} else {
			fmt.Println(err)
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
        Total_online := float64(Stats["Current_Population"])

	PoverS := Total_online / float64(Stats["Seeking_Population"])
	fmt.Printf("Total seeking : %d Ratio of pop/seeking: %1f\n", Stats["Seeking_Population"], PoverS )

	// Add PoverS to the Batch point
	WriteInflux1Tfl(conn, "Stats", "Ratio_PS", "Ratio", PoverS)

	// Average Seeking Time 
	WriteInflux1Tfl(conn, "Stats", "Seeking_Time", "Average", average_crying.Seconds())

	// Write seeking population
	WriteInflux2Tint(conn, "nasomi", "location", "Nasomi", "stat", "seeking_total", Stats["Seeking_Population"])

	// Write population
	WriteInflux2Tint(conn, "nasomi", "location", "Nasomi", "stat", "population", Stats["Current_Population"])

	// Write AH Transactions
	WriteInflux1Tfl(conn, "Stats", "Economy", "AH_Transactions", float64(Stats["AH_Transactions"]))

	// Write AH Gil in the last 24 hours 
	WriteInflux1Tfl(conn, "Stats", "Economy", "AH_gil", float64(Stats["AH_gil"]))

	// Write Mob Deaths in the last 24 hours 
	WriteInflux1Tfl(conn, "Stats", "Deaths", "mob", float64(Stats["Mob_Deaths"]))

	// Write AH Gil in the last 24 hours 
	WriteInflux1Tfl(conn, "Stats", "Deaths", "players", float64(Stats["Player_Deaths"]))
}
