package main

import (
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/warmans/stressfaktor-api/api"
	"github.com/warmans/stressfaktor-api/data"
	"github.com/warmans/stressfaktor-api/data/store"
	"github.com/warmans/stressfaktor-api/data/source/bcamp"
	"github.com/warmans/stressfaktor-api/data/source/sfaktor"
	"github.com/warmans/dbr"

	"github.com/gorilla/sessions"
)

// VERSION is used in packaging
const VERSION = "0.9.0"

func main() {

	bind := flag.String("bind", ":8080", "Web server bind address")
	terminURI := flag.String("termin", "https://stressfaktor.squat.net/termine.php?display=30", "Address of termine page")
	location := flag.String("location", "Europe/Berlin", "Time localization")
	dbPath := flag.String("dbpath", "./db.sqlite3", "Location of DB file")
	ver := flag.Bool("v", false, "Print version and exit")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	runIngest := flag.Bool("ingest", true, "Periodically ingest new data")
	authKey := flag.String("auth.key", "changeme91234567890123456789012", "key used to create sessions")

	flag.Parse()

	if *ver {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	//localize time
	time.LoadLocation(*location)

	db, err := dbr.Open("sqlite3", *dbPath, nil)
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err.Error())
	}
	defer db.Close()

	//setup database (if required)
	if err := store.InitializeSchema(db.NewSession(nil)); err != nil {
		log.Fatalf("Failed to initialize local DB: %s", err.Error())
	}

	dataStore := &store.Store{DB: db.NewSession(nil)}
	userStore := &store.UserStore{DB: db.NewSession(nil)}

	if *runIngest {
		dataIngest := data.Ingest{
			DB: db.NewSession(nil),
			UpdateFrequency: time.Duration(1) * time.Hour,
			Stressfaktor:  &sfaktor.Crawler{TermineURI: *terminURI},
			EventVisitors: []store.EventVisitor{
				&store.EventStoreVisitor{Store: dataStore},
				&store.BandcampVisitor{Bandcamp: &bcamp.Bandcamp{HTTP: http.DefaultClient}, VerboseLogging: *verbose},
			},
		}
		go dataIngest.Run()
	}

	//sessions
	if *authKey == "" {
		log.Fatal("You must specify an auth.key")
	}

	API := api.API{
		AppVersion: VERSION,
		DataStore: dataStore,
		UserStore: userStore,
		SessionStore: sessions.NewCookieStore([]byte(*authKey)),
	}
	http.Handle("/api/v1/", http.StripPrefix("/api/v1", API.NewServeMux()))

	log.Printf("API listening on %s", *bind)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
