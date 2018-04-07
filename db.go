package main

import (
        "fmt"
        "log"
	"time"
        "database/sql"
        _ "github.com/go-sql-driver/mysql"
)

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
		//returned_player.Minutes_seeking = Minutes_seeking
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
		res, err := addplayer.Exec(players[i].Name, players[i].Lastseen, players[i].Lastseen, players[i].Jobtxt, players[i].Mainlevel, players[i].Sublevel, players[i].Mainjob, players[i].Subjob)
		if err != nil {
			log.Fatal(err)
			fmt.Println(res.RowsAffected())
		}
	}
}

func DeleteMysqlPlayer(players []*Player, db *sql.DB) {
	deleteplayer, err := db.Prepare("DELETE FROM nasomi.Players_Seeking where Name=?")
	if err != nil {
                log.Fatal("Delete Player Prepare failure")
        }
	for i:= range(players) {
		res, err := deleteplayer.Exec(players[i].Name)
		if err != nil {
			fmt.Println(res.RowsAffected())
			log.Fatal("Failed while adding a player")
		}
	}
}


func GetDb(db *sql.DB, players []*Player) []*Player {
        // Get the users from the Mysql Database
        db_players := GetMysqlPlayers(db)

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