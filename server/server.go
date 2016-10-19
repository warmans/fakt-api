package server

import (
	"log"
	"net/http"
	"time"

	"fmt"

	"github.com/gorilla/sessions"
	"github.com/warmans/dbr"
	v1 "github.com/warmans/fakt-api/server/api.v1"
	"github.com/warmans/fakt-api/server/data"
	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/fakt-api/server/data/service/user"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"github.com/warmans/fakt-api/server/data/service/venue"
	"github.com/warmans/fakt-api/server/data/source"
	"github.com/warmans/fakt-api/server/data/source/k9"
	"github.com/warmans/fakt-api/server/data/source/sfaktor"
	"github.com/warmans/go-bandcamp-search/bcamp"
)

// VERSION is used in packaging
var Version string

type Config struct {
	ServerBind             string
	ServerLocation         string
	CrawlerStressfaktorURI string
	DbPath                 string
	EncryptionKey          string
	CrawlerRun             bool
	VerboseLogging         bool
}

type Server struct {
	conf *Config
}

func (s *Server) Start() error {

	//localize time
	time.LoadLocation(s.conf.ServerLocation)

	db, err := dbr.Open("sqlite3", s.conf.DbPath, nil)
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err.Error())
	}
	defer db.Close()

	//setup database (if required)
	if err := data.InitializeSchema(db.NewSession(nil)); err != nil {
		log.Fatalf("Failed to initialize local DB: %s", err.Error())
	}

	utagStore := &utag.UTagService{DB: db.NewSession(nil)}
	eventStore := &event.EventService{DB: db.NewSession(nil)}
	performerStore := &performer.PerformerService{DB: db.NewSession(nil), UTagService: utagStore}
	venueStore := &venue.VenueService{DB: db.NewSession(nil)}
	userStore := &user.UserStore{DB: db.NewSession(nil)}

	if s.conf.CrawlerRun {
		dataIngest := data.Ingest{
			DB:              db.NewSession(nil),
			UpdateFrequency: time.Duration(1) * time.Hour,
			Crawlers: []source.Crawler{
				&sfaktor.Crawler{TermineURI: s.conf.CrawlerStressfaktorURI},
				&k9.Crawler{},
			},
			EventVisitors: []common.EventVisitor{
				&data.PerformerServiceVisitor{PerformerService: performerStore},
				&data.BandcampVisitor{Bandcamp: &bcamp.Bandcamp{HTTP: http.DefaultClient}, VerboseLogging: s.conf.VerboseLogging},
			},
			EventService:     eventStore,
			PerformerService: performerStore,
			VenueService:     venueStore,
		}
		go dataIngest.Run()
	}

	//sessions
	if s.conf.EncryptionKey == "" {
		return fmt.Errorf("You must specify an auth.key")
	}

	API := v1.API{
		AppVersion:       Version,
		UserService:      userStore,
		EventService:     eventStore,
		VenueService:     venueStore,
		PerformerService: performerStore,
		UTagService:      utagStore,
		SessionStore:     sessions.NewCookieStore([]byte(s.conf.EncryptionKey)),
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", API.NewServeMux()))

	log.Printf("API listening on %s", s.conf.ServerBind)
	return http.ListenAndServe(s.conf.ServerBind, mux)
}

func NewServer(conf *Config) *Server {
	return &Server{conf: conf}
}
