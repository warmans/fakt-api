package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/warmans/stressfaktor-api/server"
)

var (
	bind      = flag.String("bind", ":8080", "Web server bind address")
	terminURI = flag.String("termin", "https://stressfaktor.squat.net/termine.php?display=30", "Address of termine page")
	location  = flag.String("location", "Europe/Berlin", "Time localization")
	dbPath    = flag.String("dbpath", "./db.sqlite3", "Location of DB file")
	ver       = flag.Bool("v", false, "Print version and exit")
	verbose   = flag.Bool("verbose", false, "Verbose logging")
	runIngest = flag.Bool("ingest", true, "Periodically ingest new data")
	authKey   = flag.String("auth.key", "changeme91234567890123456789012", "key used to create sessions")
)

func main() {

	flag.Parse()

	if *ver {
		fmt.Printf("%s", server.VERSION)
		os.Exit(0)
	}

	config := &server.Config{
		ServerBind:     *bind,
		ServerLocale:   *location,
		TermineURI:     *terminURI,
		DbPath:         *dbPath,
		RunIngest:      *runIngest,
		EncryptionKey:  *authKey,
		VerboseLogging: *verbose,
	}

	log.Fatal(server.NewServer(config).Start())
}
