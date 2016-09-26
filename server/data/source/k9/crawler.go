package k9

import (
	"fmt"
	"regexp"
	"time"

	"log"
	"strings"

	"github.com/warmans/stressfaktor-api/server/data/store"
	"github.com/ungerik/go-rss"
)

var dateRegex = regexp.MustCompile(`[^0-9]+([0-9]{2})\.([0-9]{2})\.([0-9]{4})[^0-9]+([0-9]{2})\.([0-9]{2})`)

const (
	K9Name    = "K9"
	K9FeedURI = "http://www.kinzig9.de/rss.xml"
	K9Address = "Kinzigstr. 9, 10247 Berlin"
)

type Crawler struct{}

func (c *Crawler) Crawl(localTime *time.Location) ([]*store.Event, error) {
	events := make([]*store.Event, 0)

	channel, err := rss.Read(K9FeedURI)
	if err != nil {
		return events, err
	}

	for _, itm := range channel.Item {
		if event := eventFromFeedItem(itm, localTime); event != nil {
			events = append(events, event)
		}
	}
	return events, nil
}

func eventFromFeedItem(item rss.Item, localTime *time.Location) *store.Event {

	date, err := dateFromTitle(item.Title, localTime)
	if err != nil {
		log.Print("Failed to parse date in K9 RSS: %s", err.Error())
		return nil
	}

	return &store.Event{
		Date: date,
		Venue: &store.Venue{
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
