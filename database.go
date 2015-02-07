package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var (
	table_tweets = "CREATE TABLE IF NOT EXISTS tweets (ID INTEGER PRIMARY KEY AUTOINCREMENT, URL TEXT UNIQUE NOT NULL, HEADLINE TEXT DEFAULT NULL);"
)

type Tweet struct {
	url      string `db:"url"`
	headline string `db:"headline"`
}

func dbInit() (err error) {
	db := dbConnect()
	defer dbDisconnect(db)

	log.Println("Trying to ping database")
	err = db.Ping()
	if err != nil {
		return err
	} else {
		log.Println("Database ping successful")
	}

	err = schemaInit(db)
	if err != nil {
		return err
	} else {
		log.Println("Schema creation/validation successful")
	}

	return nil
}

func dbConnect() (db *sql.DB) {
	db, err := sql.Open("sqlite3",
		DatabaseLocation)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func dbDisconnect(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func schemaInit(db *sql.DB) error {
	_, err := db.Exec(table_tweets)
	return err
}

func prepareInsertStatement(db *sql.DB) (insertStmt *sql.Stmt) {
	insertStmt, err := db.Prepare("INSERT INTO tweets (url, headline) VALUES (?,?);")
	if err != nil {
		log.Fatal(err)
	}
	return insertStmt
}

func prepareSelectStatement(db *sql.DB) (selectStmt *sql.Stmt) {
	selectStmt, err := db.Prepare("SELECT url,headline FROM tweets WHERE url = ?;")
	if err != nil {
		log.Fatal(err)
	}
	return selectStmt
}

func insertTweet(tweet Tweet) error {
	db := dbConnect()
	defer dbDisconnect(db)
	stmt := prepareInsertStatement(db)
	defer stmt.Close()

	log.Println("Inserting '" + tweet.headline + "'")
	_, err := stmt.Exec(tweet.url, tweet.headline)
	return err
}

func getTweetByUrl(url string) (tweet *Tweet, err error) {
	tweet = &Tweet{}
	db := dbConnect()
	defer dbDisconnect(db)
	stmt := prepareSelectStatement(db)
	defer stmt.Close()

	log.Println("Trying to retrieve '" + url + "'")
	err = stmt.QueryRow(url).Scan(&(tweet.url), &(tweet.headline))
	return tweet, err
}
