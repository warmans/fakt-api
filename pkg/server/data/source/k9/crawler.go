package k9

import (
	"fmt"
	"regexp"
	"time"

	"strings"

	"github.com/go-kit/kit/log"
	"github.com/ungerik/go-rss"
	"github.com/warmans/fakt-api/pkg/server/data/service/common"
)

var dateRegex = regexp.MustCompile(`[^0-9]+([0-9]{2})\.([0-9]{2})\.([0-9]{4})[^0-9]+([0-9]{2})\.([0-9]{2})`)

const (
	K9Name    = "K9"
	K9FeedURI = "http://www.kinzig9.de/rss.xml"
	K9Address = "Kinzigstr. 9, 10247 Berlin"
)

type Crawler struct {
	Timezone *time.Location
	Logger   log.Logger
}

func (c *Crawler) Name() string {
	return "k9"
}

func (c *Crawler) Crawl() ([]*common.Event, error) {
	events := make([]*common.Event, 0)

	channel, err := rss.Read(K9FeedURI)
	if err != nil {
		return events, err
	}

	for _, itm := range channel.Item {
		if event := c.eventFromFeedItem(itm, c.Timezone); event != nil {
			events = append(events, event)
		}
	}
	return events, nil
}

func (c *Crawler) eventFromFeedItem(item rss.Item, localTime *time.Location) *common.Event {

	date, err := dateFromTitle(item.Title, localTime)
	if err != nil {
		c.Logger.Log("msg", fmt.Sprintf("Failed to parse date in K9 RSS: %s", err.Error()))
		return nil
	}

	return &common.Event{
		Date: date,
		Venue: &common.Venue{
			Name:    K9Name,
			Address: K9Address,
		},
		Description: item.Description,
	}
}

func dateFromTitle(title string, localTime *time.Location) (time.Time, error) {
	matches := dateRegex.FindAllStringSubmatch(title, 5)
	if len(matches) == 0 {
		return time.Time{}, fmt.Errorf("No date found in title (%s)", title)
	}
	if len(matches[0]) != 6 {
		return time.Time{}, fmt.Errorf("Wrong number of items returned from title (%s): %s", title, strings.Join(matches[0], ","))
	}
	return time.ParseInLocation("02-01-2006 15:04", fmt.Sprintf("%s-%s-%s %s:%s", matches[0][1], matches[0][2], matches[0][3], matches[0][4], matches[0][5]), localTime)
}
