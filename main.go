package main

import (
	"flag"
	"net/http"
	"log"
	"github.com/warmans/stressfaktor-api/crawler"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
	"github.com/warmans/stressfaktor-api/entity"
	"encoding/json"
	"os"
	"fmt"
)

const VERSION = "0.0.3"

type Response struct {
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
}

func SendResponse(rw http.ResponseWriter, response *Response) {
	rw.Header().Add("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(response.Status)

	jsonEncoder := json.NewEncoder(rw)
	jsonEncoder.Encode(response)
}

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

	scraper := &crawler.Crawler{EventStore: eventStore, TermineURI: *terminURI}
	scraper.Run(time.Duration(1) * time.Hour)

	http.HandleFunc("/api/v1/event", func(rw http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()
		r.ParseForm()

		events, err := eventStore.Find(&entity.EventFilter{})
		if err != nil {
			log.Print(err.Error())
			http.Error(rw, "Failed", http.StatusInternalServerError)
			return
		}

		SendResponse(rw, &Response{Status: 200, Payload: events})
	})

	http.HandleFunc("/api/v1/version", func(rw http.ResponseWriter, r *http.Request) {
		SendResponse(rw, &Response{Status: 200, Payload: map[string]string{"version": VERSION}})
	})

	log.Fatal(http.ListenAndServe(*bind, nil))
}
