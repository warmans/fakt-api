package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

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
		DbPath:                 *dbPath,
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

	logger.Fatal("Server Exited", zap.Error(server.NewServer(config, logger).Start()))
}
