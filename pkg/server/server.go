package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/warmans/dbr"
	v1 "github.com/warmans/fakt-api/pkg/server/api.v1"
	"github.com/warmans/fakt-api/pkg/server/data"
	"github.com/warmans/fakt-api/pkg/server/data/media"
	"github.com/warmans/fakt-api/pkg/server/data/process"
	"github.com/warmans/fakt-api/pkg/server/data/source"
	"github.com/warmans/fakt-api/pkg/server/data/source/k9"
	"github.com/warmans/fakt-api/pkg/server/data/source/sfaktor"
	"github.com/warmans/fakt-api/pkg/server/data/store/common"
	"github.com/warmans/fakt-api/pkg/server/data/store/event"
	"github.com/warmans/fakt-api/pkg/server/data/store/performer"
	"github.com/warmans/fakt-api/pkg/server/data/store/tag"
	"github.com/warmans/fakt-api/pkg/server/data/store/venue"
	"github.com/warmans/go-bandcamp-search/bcamp"
	"go.uber.org/zap"
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

func NewServer(conf *Config, logger *zap.Logger) *Server {
	return &Server{conf: conf, logger: logger}
}

type Server struct {
	conf   *Config
	logger *zap.Logger
}

func (s *Server) Start() error {

	//localize time
	time.LoadLocation(s.conf.ServerLocation)

	db, err := dbr.Open("sqlite3", s.conf.DbPath, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	//setup database (if required)
	if err := data.InitializeSchema(db.NewSession(nil)); err != nil {
		return errors.Wrap(err, "Failed to initialize local DB")
	}

	performerService := &performer.Store{DB: db.NewSession(nil), Logger: s.logger}
	eventService := &event.Store{DB: db.NewSession(nil), PerformerService: performerService}
	venueService := &venue.Store{DB: db.NewSession(nil)}
	tagService := &tag.Store{DB: db.NewSession(nil)}

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
				&sfaktor.Crawler{
					TermineURI: s.conf.CrawlerStressfaktorURI,
					Timezone: tz,
					Logger: s.logger.With(zap.String("component", "sfaktor crawler")),
				},
				&k9.Crawler{
					Timezone: tz,
					Logger: s.logger.With(zap.String("component", "k9crawler")),
				},
			},
			EventVisitors: []common.EventVisitor{
				&data.PerformerServiceVisitor{PerformerService: performerService, Logger: s.logger},
				&data.BandcampVisitor{Bandcamp: &bcamp.Bandcamp{HTTP: http.DefaultClient},Logger: s.logger, ImageMirror: imageMirror},
			},
			EventService:     eventService,
			PerformerService: performerService,
			VenueService:     venueService,
			Logger:           s.logger.With(zap.String("component", "ingest")),
		}
		go dataIngest.Run()

		//pre-calculate some stats when ingest is running

		//performer activity
		activityRunner := process.GetActivityRunner(time.Minute*10, s.logger)
		go activityRunner.Run(db.NewSession(nil))
	}

	//sessions
	if s.conf.EncryptionKey == "" {
		return fmt.Errorf("you must specify an auth.key")
	}

	API := v1.API{
		AppVersion:       Version,
		EventService:     eventService,
		VenueService:     venueService,
		PerformerService: performerService,
		TagService:       tagService,
		SessionStore:     sessions.NewCookieStore([]byte(s.conf.EncryptionKey)),
		Logger:           s.logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", API.NewServeMux()))

	staticFileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", gziphandler.GzipHandler(staticFileServer)))

	s.logger.Info(fmt.Sprintf("API listening on %s", s.conf.ServerBind))
	return http.ListenAndServe(s.conf.ServerBind, mux)
}
