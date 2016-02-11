package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/warmans/stressfaktor-api/crawler"
	"github.com/warmans/stressfaktor-api/entity"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/warmans/stressfaktor-api/api"
	"github.com/warmans/stressfaktor-api/data/bcamp"
)

// VERSION is used in packaging
const VERSION = "0.2.0"

func main() {

	bind := flag.String("bind", ":8080", "Web server bind address")
	terminURI := flag.String("termin", "https://stressfaktor.squat.net/termine.php?display=90", "Address of termine page")
	location := flag.String("location", "Europe/Berlin", "Time localization")
	dbPath := flag.String("dbpath", "./db.sqlite3", "Location of DB file")
	ver := flag.Bool("v", false, "Print version and exit")
	flag.Parse()

	if *ver {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	//localize time
	time.LoadLocation(*location)

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err.Error())
	}
	defer db.Close()

	eventStore := &entity.EventStore{DB: db}
	if err := eventStore.Initialize(); err != nil {
		log.Fatalf("Failed to initialize local DB: %s", err.Error())
	}

	scraper := &crawler.Crawler{
		EventStore: eventStore,
		TermineURI: *terminURI,
		Bandcamp: &bcamp.Bandcamp{HTTP: http.DefaultClient},
	}
	go scraper.Run(time.Duration(1) * time.Hour)

	API := api.API{AppVersion: VERSION, EventStore: eventStore}
	http.Handle("/api/v1/", http.StripPrefix("/api/v1", API.NewServeMux()))

	log.Printf("API listening on %s", *bind)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
