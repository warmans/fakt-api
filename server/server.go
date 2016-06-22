package server

import (
	"log"
	"net/http"
	"time"

	"fmt"

	"github.com/gorilla/sessions"
	"github.com/warmans/dbr"
	"github.com/warmans/go-bandcamp-search/bcamp"
	"github.com/warmans/stressfaktor-api/server/api"
	"github.com/warmans/stressfaktor-api/server/data"
	"github.com/warmans/stressfaktor-api/server/data/source/sfaktor"
	"github.com/warmans/stressfaktor-api/server/data/store"
)

// VERSION is used in packaging
var Version string

type Config struct {
	ServerBind     string
	ServerLocale   string
	TermineURI     string
	DbPath         string
	EncryptionKey  string
	RunIngest      bool
	VerboseLogging bool
}

type Server struct {
	conf *Config
}

func (s *Server) Start() error {

	//localize time
	time.LoadLocation(s.conf.ServerLocale)

	db, err := dbr.Open("sqlite3", s.conf.DbPath, nil)
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

	if s.conf.RunIngest {
		dataIngest := data.Ingest{
			DB:              db.NewSession(nil),
			UpdateFrequency: time.Duration(1) * time.Hour,
			Stressfaktor:    &sfaktor.Crawler{TermineURI: s.conf.TermineURI},
			EventVisitors: []store.EventVisitor{
				&store.EventStoreVisitor{Store: dataStore},
				&store.BandcampVisitor{Bandcamp: &bcamp.Bandcamp{HTTP: http.DefaultClient}, VerboseLogging: s.conf.VerboseLogging},
			},
		}
		go dataIngest.Run()
	}

	//sessions
	if s.conf.EncryptionKey == "" {
		return fmt.Errorf("You must specify an auth.key")
	}

	API := api.API{
		AppVersion:   Version,
		DataStore:    dataStore,
		UserStore:    userStore,
		SessionStore: sessions.NewCookieStore([]byte(s.conf.EncryptionKey)),
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", API.NewServeMux()))

	log.Printf("API listening on %s", s.conf.ServerBind)
	return http.ListenAndServe(s.conf.ServerBind, mux)
}

func NewServer(conf *Config) *Server {
	return &Server{conf: conf}
}
