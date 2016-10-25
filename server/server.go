package server

import (
	"net/http"
	"time"

	"fmt"

	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/sessions"
	"github.com/warmans/dbr"
	v1 "github.com/warmans/fakt-api/server/api.v1"
	"github.com/warmans/fakt-api/server/data"
	"github.com/warmans/fakt-api/server/data/media"
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
	StaticFilesPath        string
	VerboseLogging         bool
}

func NewServer(conf *Config, logger log.Logger) *Server {
	return &Server{conf: conf, logger: logger}
}

type Server struct {
	conf   *Config
	logger log.Logger
}

func (s *Server) Start() error {

	//localize time
	time.LoadLocation(s.conf.ServerLocation)

	db, err := dbr.Open("sqlite3", s.conf.DbPath, nil)
	if err != nil {
		s.logger.Log("msg", fmt.Sprintf("Failed to open DB: %s", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	//setup database (if required)
	if err := data.InitializeSchema(db.NewSession(nil)); err != nil {
		s.logger.Log("msg", fmt.Sprintf("Failed to initialize local DB: %s", err.Error()))
		os.Exit(1)
	}

	utagService := &utag.UTagService{DB: db.NewSession(nil)}
	performerService := &performer.PerformerService{DB: db.NewSession(nil), UTagService: utagService, Logger: s.logger}
	eventService := &event.EventService{DB: db.NewSession(nil), UTagService: utagService, PerformerService: performerService}
	venueService := &venue.VenueService{DB: db.NewSession(nil)}
	userService := &user.UserStore{DB: db.NewSession(nil)}

	imageMirror := media.NewImageMirror(s.conf.StaticFilesPath)

	if s.conf.CrawlerRun {
		tz, err := source.MustMakeTimeLocation("Europe/Berlin")
		if err != nil {
			panic(err.Error())
		}

		dataIngest := data.Ingest{
			DB:              db.NewSession(nil),
			UpdateFrequency: time.Duration(1) * time.Hour,
			Crawlers: []source.Crawler{
				&sfaktor.Crawler{TermineURI: s.conf.CrawlerStressfaktorURI, Timezone: tz},
				&k9.Crawler{Timezone: tz, Logger: log.NewContext(s.logger).With("component", "k9crawler")},
			},
			EventVisitors: []common.EventVisitor{
				&data.PerformerServiceVisitor{PerformerService: performerService, Logger: s.logger},
				&data.BandcampVisitor{Bandcamp: &bcamp.Bandcamp{HTTP: http.DefaultClient}, VerboseLogging: s.conf.VerboseLogging, Logger: s.logger, ImageMirror: imageMirror},
			},
			EventService:     eventService,
			PerformerService: performerService,
			VenueService:     venueService,
			Logger:           log.NewContext(s.logger).With("component", "ingest"),
		}
		go dataIngest.Run()
	}

	//sessions
	if s.conf.EncryptionKey == "" {
		return fmt.Errorf("You must specify an auth.key")
	}

	API := v1.API{
		AppVersion:       Version,
		UserService:      userService,
		EventService:     eventService,
		VenueService:     venueService,
		PerformerService: performerService,
		UTagService:      utagService,
		SessionStore:     sessions.NewCookieStore([]byte(s.conf.EncryptionKey)),
		Logger:           s.logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", API.NewServeMux()))

	staticFileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", gziphandler.GzipHandler(staticFileServer)))

	s.logger.Log("msg", fmt.Sprintf("API listening on %s", s.conf.ServerBind))
	return http.ListenAndServe(s.conf.ServerBind, mux)
}
