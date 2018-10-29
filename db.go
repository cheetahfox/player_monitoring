package main

import (
        "fmt"
        "log"
	"time"
	"os"
        "database/sql"
	"encoding/json"
        _ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/influxdb/client/v2"
)


func ConnectInfluxdb() client.Client {
	// Open the influx Db conneciton
        if os.Getenv("INFLUX_ADDRESS") == "" {
                log.Fatal("No InfluxDb addess set")
        }
        Conn, err := client.NewHTTPClient(client.HTTPConfig{
                Addr:     os.Getenv("INFLUX_ADDRESS"),
                Username: os.Getenv("INFLUX_USERNAME"),
                Password: os.Getenv("INFLUX_PASSWORD"),
        })
        if err != nil {
                log.Fatal(err)
        }
	return Conn
}

func ConnectMySql() *sql.DB {
        /*
	This function opens the MySql DB
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

        err = db.Ping()
        if err != nil {
                panic(err.Error())
        }
	return db
}

func GetMysqlTods(db *sql.DB) []*Tod {
	// Get the Current ToD's out of the database
	var (
		NM string
		Killer string
		Ls string
		First_Seen time.Time
		Last_Seen time.Time
	)
	tods := []*Tod{}

        rows, err := db.Query("SELECT * FROM nasomi.tod")
        if err != nil {
                log.Fatal(err)
        }
        defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&NM, &Killer, &Ls, &First_Seen, &Last_Seen)
		if err != nil {
			log.Fatal(err)
		}
		db_tod           := new(Tod)
		db_tod.NM         = NM
		db_tod.Killer     = Killer
		db_tod.LinkShell  = Ls
		db_tod.First_Seen = First_Seen
		db_tod.Last_Seen  = Last_Seen
		tods = append(tods, db_tod)
	}

	return tods
}

func GetMysqlPlayers(db *sql.DB) []*Player {
	var (
		Name string
		Job_txt string
		Started_seeking time.Time
		Last_seen time.Time
		//Minutes_seeking int
		Mainjob_L int
		Subjob_L int
		Main string
		Sub string
	)
	db_players := []*Player{}
        /*
        This functtion is connects to the mysql database to get players who have been seen before
        */

	rows, err := db.Query("SELECT * FROM nasomi.Players_Seeking")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&Name, &Started_seeking, &Last_seen, &Job_txt, &Mainjob_L, &Subjob_L, &Main, &Sub)
		if err != nil {
			log.Fatal(err)
		}
		returned_player := new(Player)
		returned_player.Name            = Name
		returned_player.Jobtxt         = Job_txt
		returned_player.Started_seeking = Started_seeking
		returned_player.Lastseen        = Last_seen
		returned_player.Mainjob         = Main
		returned_player.Subjob          = Sub
		returned_player.Mainlevel       = Mainjob_L
		returned_player.Sublevel        = Subjob_L
		db_players = append(db_players, returned_player)
	}

	return db_players
}

func UpdateMysqlSeen(players []*Player, db *sql.DB) {
	/*
	This function will update a players seen-time. This is the only place where I updat seen-time 
	unless am I add
	*/
	commit, err := db.Prepare("UPDATE Players_Seeking set Last_seen=? where Name=?")
	if err != nil {
		log.Fatal("MySql Prepare failure")
	}
	for i:= range(players) {
		res, err := commit.Exec(players[i].Lastseen, players[i].Name)
		if err != nil {
			fmt.Println(res.RowsAffected())
			log.Fatal("MySql exec error")
		}
	}
}

func UpdateMysqlTod(tods []*Tod, db *sql.DB) {
	// This function updates last seen on Tod's in the DB
	commit, err := db.Prepare("UPDATE nasomi.tod set last_seen=? where nm=?")
	if err != nil {
		log.Fatal(err)
	}
	for i:= range(tods) {
		res, err := commit.Exec(tods[i].Last_Seen, tods[i].NM)
		if err != nil {
			log.Println(res.RowsAffected())
			log.Fatal(err)
		}
	}
}

func AddMysqlPlayer(players []*Player, db *sql.DB) {
	/*
	This function adds a new user to the 
	*/
	addplayer, err := db.Prepare("INSERT INTO nasomi.Players_Seeking set Name=?, Started_seeking=?, Last_seen=?, Job_txt=?, Mainjob_Level=?, Subjob_Level=?, Mainjob=?, Subjob=?")
	if err != nil {
		log.Fatal("Addplay Prepare failure")
	}
	for i:= range(players) {
		// Since this is the first time, Started seeking is going to be the same as lastseen
		_, err := addplayer.Exec(players[i].Name, players[i].Lastseen, players[i].Lastseen, players[i].Jobtxt, players[i].Mainlevel, players[i].Sublevel, players[i].Mainjob, players[i].Subjob)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func AddMysqlTod(tods []*Tod, db *sql.DB) {
	/*
	This function adds new Tods to the Db
	*/

	addtod, err := db.Prepare("INSERT INTO nasomi.tod set nm=?, killer=?, ls=?, first_seen=?, last_seen=?")
	if err != nil {
		log.Fatal(err)
	}
	for i:= range(tods) {
		_, err := addtod.Exec(tods[i].NM, tods[i].Killer, tods[i].LinkShell, tods[i].Last_Seen, tods[i].Last_Seen)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func LogSeekingSession(session *Player, db *sql.DB) {
	/*
	This function creates a permanet log of each seeking session
	this will be used in the future to do long term analysis of trends. 
	*/
	logsession, err := db.Prepare("INSERT INTO nasomi.Log_Players_Seeking set Name=?, Started_seeking=?, Last_seen=?, Mainjob_Level=?, Subjob_Level=?, Mainjob=?, Subjob=?")
	if err != nil {
		fmt.Println(err)
		log.Fatal("Log Prepare Failure")
	}
	res, err := logsession.Exec(session.Name, session.Started_seeking, session.Lastseen, session.Mainlevel, session.Sublevel, session.Mainjob, session.Subjob)
	if err != nil {
		fmt.Println(res.RowsAffected())
		log.Fatal("Failure to insert into long-term log")
	}
}

func DeleteMysqlPlayer(players []*Player, db *sql.DB) {
	deleteplayer, err := db.Prepare("DELETE FROM nasomi.Players_Seeking where Name=?")
	if err != nil {
                log.Fatal("Delete Player Prepare failure")
        }
	for i:= range(players) {
		// Log the session before deleting it
		LogSeekingSession(players[i], db)
		res, err := deleteplayer.Exec(players[i].Name)
		if err != nil {
			fmt.Println(res.RowsAffected())
			log.Fatal("Failed while adding a player")
		}
	}
}

func LogTods(tod *Tod, db *sql.DB) {
	/*
	Log each Tod into a full log for later Data processing
	*/
	logtod, err := db.Prepare("INSERT INTO nasomi.log_tod set nm=?, killer=?, ls=?, first_seen=?, last_seen=?")
	if err != nil {
		log.Fatal(err)
	}

	_, err = logtod.Exec(tod.NM, tod.Killer, tod.LinkShell, tod.First_Seen, tod.Last_Seen)
	if err != nil {
		log.Fatal(err)
	}
}

func DeleteMysqlTod(tods []*Tod, db *sql.DB) {
	// Delete Tod's from the DB if they are no longer on the page
	deletetod, err := db.Prepare("DELETE FROM nasomi.tod where nm=? AND killer=?")
	if err != nil {
		log.Fatal(err)
	}
	for i:= range(tods) {
		LogTods(tods[i], db)
		_, err := deletetod.Exec(tods[i].NM, tods[i].Killer)
		if err != nil {
			log.Fatal(err)
		}
	}
}


func GetDb(db *sql.DB, players []*Player) []*Player {
	/*
	This funciton returns a cleaned up list of players in the persistent DB.
	I take in the current scraped list of players because this function also 
	performs all of the maintance on the DB. It updates players seek times.
	Delets them from the DB if they aren't seeking on that job anymore.
	Adds them if they are a new player and deletes them (which resets the seeking time)
	if they are seeking for more than 8 hours. 
	*/
        // Get the users from the Mysql Database
        db_players := GetMysqlPlayers(db)

	// Check for AFK players, we delete any players that have been seeking for more than 8 hours
	// If they get deleted here, they will be readded, this should keep someone who is afk from messing with the average time
	afkplayers := []*Player{}
	now := time.Now()
	dur, _ := time.ParseDuration("8h")
	// invert the duration to be a -8 hour duration
	if dur > 0 {
		dur = -dur
	}

	afk_time := now.Add(dur)
        for i:= range(db_players) {
		if db_players[i].Started_seeking.Before(afk_time)  {
			afkplayers = append(afkplayers, db_players[i])
		}
        }
	if afkplayers != nil {
		DeleteMysqlPlayer(afkplayers, db)
	}

        // Check if the current players are in the database, update them if they are
        updateseen_players := []*Player{}
        for i:= range(players) {
                if PlayerinDB(players[i], db_players) == true {
                        updateseen_players = append(updateseen_players, players[i])
                }
        }
        if updateseen_players != nil {
                UpdateMysqlSeen(updateseen_players, db)
        }

        // Check if the Database contains players that are not currently seeking, delete them from the DB if present
        deleteplayers := []*Player{}
        for i:= range(db_players) {
                if PlayerinDB(db_players[i], players) != true {
                        deleteplayers = append(deleteplayers, db_players[i])
                }
        }
        if deleteplayers != nil {
                DeleteMysqlPlayer(deleteplayers, db)
        }

        // Check if current players are not in the database, add them to the db if they aren't
        addplayers := []*Player{}
        for i:= range(players) {
                if PlayerinDB(players[i], db_players) != true {
                        addplayers = append(addplayers, players[i])
                }
        }
        if addplayers != nil {
                AddMysqlPlayer(addplayers, db)
        }

	// Refresh the now current list of players.
	db_players = GetMysqlPlayers(db)

	return db_players
}

func GenTodDb(db *sql.DB, scraped_tod []*Tod) {
	// similar to GetDb but for the Tod's
	// Get the current Tods
	db_tods := GetMysqlTods(db)

	// Update lastseen if the ToD is already in the DB
	NMtoUpdate := []*Tod{}
	for i:= range(scraped_tod) {
		if TodinDB(scraped_tod[i],db_tods) == true {
			NMtoUpdate = append(NMtoUpdate, scraped_tod[i])
		}
	}
	if NMtoUpdate != nil {
		UpdateMysqlTod(NMtoUpdate, db)
	}

	/*
	Delete Any Tod that doesn't isn't on the Website
	This has to happen before we insert a new Tod
	*/
	NMtoDelete := []*Tod{}
	for i:= range(db_tods) {
		if TodinDB(db_tods[i], scraped_tod) != true {
			NMtoDelete = append(NMtoDelete, db_tods[i])
		}
	}
	if NMtoDelete != nil {
		// This fuction Logs and Deletes the Tod from the DB
		DeleteMysqlTod(NMtoDelete, db)
	}

	/*
	Check if it's a new Tod, If so add them to the DB
	*/
	TodtoAdd := []*Tod{}
	for i:= range(scraped_tod) {
		if TodinDB(scraped_tod[i], db_tods) != true {
			TodtoAdd = append(TodtoAdd, scraped_tod[i])
		}
	}
	if TodtoAdd != nil {
		AddMysqlTod(TodtoAdd, db)
	}
}

func WriteInflux1Tfl(conn client.Client, measure string, tag1 string, tag2 string, value float64) {
	/*
	This function takes a single float64 value and writes it into the InfluxDb with two tags.
	This is useful since I often only want to write a single value.
	*/
        bp, err := client.NewBatchPoints(client.BatchPointsConfig{
                Database: os.Getenv("INFLUX_DB"),
                Precision: "s",
        })
        if err != nil {
                log.Fatal(err)
        }

        tags := map[string]string{tag1:tag2}
        fields := map[string]interface{}{
                "value": value,
        }
        pt, err := client.NewPoint(measure, tags, fields, time.Now())
        if err != nil {
                log.Fatal(err)
        }
        bp.AddPoint(pt)

        if err := conn.Write(bp); err != nil {
                log.Fatal(err)
        }
}

func GetNasomiPop(conn client.Client ) float64 {
        /*
        Get the toal number of people online from the Influxdb database
        */
	var Total_pop float64
        q := client.NewQuery("Select last(\"value\") FROM \"nasomi\" WHERE(\"location\" = 'Nasomi' AND \"stat\" = 'population')", "nasomi", "s")
        if response, err := conn.Query(q); err == nil && response.Error() == nil {
                data, err :=  response.Results[0].Series[0].Values[0][1].(json.Number).Float64()
                if err != nil {
                        log.Fatal("json.number failed in influxdb response")
                }
		Total_pop = data
        } else {
                log.Fatal(response.Error(), err)
        }
	return Total_pop
}

func WriteInflux2Tint(conn client.Client, measure string, tag string, tag_value string, tag2 string, tag2_value string, int_value int) {
	/*
	This function takes a single float64 value and writes it into the InfluxDb with two tags.
	This is useful since I often only want to write a single value.
	*/
	value := float64(int_value)
        bp, err := client.NewBatchPoints(client.BatchPointsConfig{
                Database: os.Getenv("INFLUX_DB"),
                Precision: "s",
        })
        if err != nil {
                log.Fatal(err)
        }

        tags := map[string]string{tag:tag_value,tag2:tag2_value}
        fields := map[string]interface{}{
                "value": value,
        }
        pt, err := client.NewPoint(measure, tags, fields, time.Now())
        if err != nil {
                log.Fatal(err)
        }
        bp.AddPoint(pt)

        if err := conn.Write(bp); err != nil {
                log.Fatal(err)
        }
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

func TodinDB( tod *Tod, db []*Tod) (bool) {
	/*
	This function checks to see if a ToD is in the different slice of Tod's
	*/
	for index := range(db) {
		if tod.NM == db[index].NM {
			if tod.Killer == db[index].Killer {
				return true
			}
		}
	}
	return false
}

