package source

import (
	"time"

	"fmt"

	"github.com/warmans/fakt-api/server/data/service/common"
)

type Crawler interface {
	Crawl() ([]*common.Event, error)
	Name() string
}

func MustMakeTimeLocation(locationName string) (*time.Location, error) {
	var err error
	localTime, err := time.LoadLocation(locationName)
	if err != nil {
		return nil, fmt.Errorf("Failed to create time location %s because %s", locationName, err.Error())
	}
	return localTime, nil
}
