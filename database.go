package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	table_tweets = "CREATE TABLE IF NOT EXISTS tweets (ID INTEGER PRIMARY KEY AUTOINCREMENT, URL TEXT UNIQUE NOT NULL, HEADLINE TEXT DEFAULT NULL, GUID TEXT UNIQUE NOT NULL);"
)

type Tweet struct {
	url      string `db:"url"`
	headline string `db:"headline"`
	guid     string `db:"guid"`
}

func dbInit() (err error) {
	db := dbConnect()
	defer dbDisconnect(db)

	logger.Println("Trying to ping database")
	err = db.Ping()
	if err != nil {
		return err
	} else {
		logger.Println("Database ping successful")
	}

	err = schemaInit(db)
	if err != nil {
		return err
	} else {
		logger.Println("Schema creation/validation successful")
	}

	return nil
}

func dbConnect() (db *sql.DB) {
	db, err := sql.Open("sqlite3",
		DatabaseLocation)
	if err != nil {
		logger.Fatal(err)
	}
	return db
}

func dbDisconnect(db *sql.DB) {
	err := db.Close()
	if err != nil {
		logger.Fatal(err)
	}
}

func schemaInit(db *sql.DB) error {
	_, err := db.Exec(table_tweets)
	return err
}

func prepareInsertStatement(db *sql.DB) (insertStmt *sql.Stmt) {
	insertStmt, err := db.Prepare("INSERT INTO tweets (url, headline, guid) VALUES (?,?,?);")
	if err != nil {
		logger.Fatal(err)
	}
	return insertStmt
}

func prepareSelectStatement(db *sql.DB) (selectStmt *sql.Stmt) {
	selectStmt, err := db.Prepare("SELECT url,headline,guid FROM tweets WHERE guid = ?;")
	if err != nil {
		logger.Fatal(err)
	}
	return selectStmt
}

func prepareCleanupStatement(db *sql.DB) (cleanupStmt *sql.Stmt) {
	cleanupStmt, err := db.Prepare("DELETE FROM tweets WHERE ID NOT IN (SELECT ID FROM tweets ORDER BY ID DESC LIMIT ?)")
	if err != nil {
		logger.Fatal(err)
	}
	return
}

func databaseCleanup() error {
	db := dbConnect()
	defer dbDisconnect(db)
	stmt := prepareCleanupStatement(db)
	defer stmt.Close()
	logger.Println("Performing Database Cleanup")
	_, err := stmt.Exec(CacheSize)
	return err
}

func insertTweet(tweet Tweet) error {
	db := dbConnect()
	defer dbDisconnect(db)
	stmt := prepareInsertStatement(db)
	defer stmt.Close()

	logger.Println("\tCaching '" + tweet.url + "'")
	_, err := stmt.Exec(tweet.url, tweet.headline, tweet.guid)
	return err
}

func getTweetByGuid(guid string) (tweet *Tweet, err error) {
	tweet = &Tweet{}
	db := dbConnect()
	defer dbDisconnect(db)
	stmt := prepareSelectStatement(db)
	defer stmt.Close()

	logger.Println("\tProbing cache for '" + guid + "'")
	err = stmt.QueryRow(guid).Scan(&(tweet.url), &(tweet.headline), &(tweet.guid))
	return tweet, err
}
