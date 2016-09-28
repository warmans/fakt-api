package source

import (
	"time"

	"github.com/warmans/fakt-api/server/data/store"
	"log"
)

type Crawler interface {
	Crawl(localTime *time.Location) ([]*store.Event, error)
	Name() string
}

func MustMakeTimeLocation(locationName string) *time.Location {
	var err error
	localTime, err := time.LoadLocation(locationName)
	if err != nil {
		log.Printf("Cannot load localtime (%s). Event times may be wrong.", err.Error())
	}
	return localTime
}