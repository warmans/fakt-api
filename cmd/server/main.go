package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rubenv/sql-migrate"
	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/pkg/server"
	"go.uber.org/zap"
)

var (
	serverBind             = flag.String("server.bind", ":8080", "Web server bind address")
	serverEncryptionKey    = flag.String("server.encryption.key", "changeme91234567890123456789012", "Key used to create sessions")
	crawlerStressfaktorURI = flag.String("crawler.stressfaktor.uri", "https://stressfaktor.squat.net/termine.php?display=30", "Address of termine page")
	crawlerLocation        = flag.String("crawler.location", "Europe/Berlin", "Time localization")
	crawlerRun             = flag.Bool("crawler.run", true, "Periodically ingest new data")
	dbPath                 = flag.String("db.path", "./db.sqlite3", "Location of DB file")
	verbose                = flag.Bool("log.verbose", false, "Verbose logging")
	staticFilesPath        = flag.String("static.path", "static", "Location to store static files")
	migrationsPath         = flag.String("migrations.path", "migrations", "Location of migrations")
	migrationsDisabled     = flag.Bool("migrations.disabled", false, "Skip applying migrations")
	ver                    = flag.Bool("v", false, "Print version and exit")
)

func main() {

	flag.Parse()

	if *ver {
		fmt.Printf("%s", server.Version)
		os.Exit(0)
	}

	config := &server.Config{
		ServerBind:             *serverBind,
		ServerLocation:         *crawlerLocation,
		CrawlerStressfaktorURI: *crawlerStressfaktorURI,
		CrawlerRun:             *crawlerRun,
		EncryptionKey:          *serverEncryptionKey,
		VerboseLogging:         *verbose,
		StaticFilesPath:        *staticFilesPath,
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("Failed to create logger")
		os.Exit(1)
	}
	defer logger.Sync()

	db, err := dbr.Open("sqlite3", *dbPath, nil)
	if err != nil {
		logger.Fatal("Failed to open DB", zap.Error(err))
	}
	defer db.Close()

	// apply migrations on startup
	if !*migrationsDisabled {
		n, err := migrate.Exec(db.DB, "sqlite3", &migrate.FileMigrationSource{Dir: *migrationsPath}, migrate.Up)
		if err != nil {
			logger.Fatal("Migrations failed", zap.Error(err))
		}
		logger.Info("Applied  migrations", zap.Int("num", n))
	}

	logger.Fatal("Server Exited", zap.Error(server.NewServer(config, logger, db).Start()))
}
