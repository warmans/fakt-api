package source

import (
	"fmt"
	"time"

	"github.com/warmans/fakt-api/pkg/server/data/store/common"
)

type Crawler interface {
	Crawl() ([]*common.Event, error)
	Name() string
}

func MustMakeTimeLocation(locationName string) (*time.Location, error) {
	var err error
	localTime, err := time.LoadLocation(locationName)
	if err != nil {
		return nil, fmt.Errorf("failed to create time location %s because %s", locationName, err.Error())
	}
	return localTime, nil
}
