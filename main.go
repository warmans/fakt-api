package main

import (
	"flag"
	"net/http"
	"log"
	"github.com/warmans/stressfaktor-api/crawler"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
	"github.com/warmans/stressfaktor-api/store"
)

func main() {

	bind := flag.String("bind", "localhost:8080", "Web server bind address")
	terminURI := flag.String("termin", "https://stressfaktor.squat.net/termine.php?display=90", "Address of termine page")
	location := flag.String("location", "Europe/Berlin", "Time localization")

	//localize time
	time.LoadLocation(*location)

	db, err := sql.Open("sqlite3", "./db.sqlite3")
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err.Error())
	}
	defer db.Close()

	eventStore := &store.EventStore{DB: db}
	if err := eventStore.Initialize(); err != nil {
		log.Fatalf("Failed to initialize local DB: %s", err.Error())
	}

	scraper := &crawler.Crawler{EventStore: eventStore, TermineURI: *terminURI}
	scraper.Run(time.Duration(1) * time.Hour)

	log.Fatal(http.ListenAndServe(*bind, nil))
}
