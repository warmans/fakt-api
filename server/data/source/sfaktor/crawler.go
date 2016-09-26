package sfaktor

import (
	"bytes"
	"errors"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	"github.com/warmans/stressfaktor-api/server/data/store"
)

var validDate = regexp.MustCompile(`^[A-Za-z]+, [0-9]{2}\.[0-9]{2}\.[0-9]{4}$`)
var validTime = regexp.MustCompile(`^[0-9]{2}\.[0-9]{2}$`)

type Crawler struct {
	TermineURI string
}

func (c *Crawler) Crawl(localTime *time.Location) ([]*store.Event, error) {

	events := make([]*store.Event, 0)

	res, err := http.Get(c.TermineURI)
	if err != nil {
		return events, fmt.Errorf("SF crawler fetch failed: %s", err.Error())
	}
	defer res.Body.Close()

	//must convert to utf-8 or the special chars will be broken
	utfBody, err := iconv.NewReader(res.Body, "ISO-8859-1", "utf-8")
	if err != nil {
		return events, fmt.Errorf("SF crawler failed to convert character set: %s", err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		return events, fmt.Errorf("SF crawler failed to parse result: %s", err.Error())
	}

	//select the main data column and handle all the sub-tables
	doc.Find("body > table:nth-child(4) > tbody > tr > td:nth-child(2) > table").Each(func(i int, sel *goquery.Selection) {
		events = append(events, c.HandleDateTable(i, sel, localTime)...)
	})

	return events, nil
}

func (c *Crawler) HandleDateTable(i int, sel *goquery.Selection, localTime time.Location) []*store.Event {

	var dateStr string
	var time time.Time
	var failed bool
	var events []*store.Event

	sel.Find("tr").Each(func(i int, tr *goquery.Selection) {

		if failed {
			return
		}

		//first row is always the date
		if dateStr == "" {
			dateStr = strings.TrimSpace(tr.Find("td > span").First().Text())
			if !validDate.MatchString(dateStr) {
				log.Printf("Invalid date string: %s", dateStr)
				failed = true
			}
			return //move on
		}

		//each row has a time
		timeStr := strings.TrimSpace(tr.Find("td > span").First().Text())
		if validTime.MatchString(timeStr) {
			log.Printf("Invalid time string: %s", timeStr)
			return //don't fail
		}

		var err error
		time, err = ParseTime(dateStr, timeStr, localTime)
		if err != nil {
			log.Printf("Failed to parse date: %s %s (%s)", dateStr, timeStr, err.Error())
			return
		}

		event, err := c.CreateEvent(time, tr.Find("td").Last())
		if err != nil {
			log.Printf("Failed to create event: %s", err.Error())
			return
		}

		events = append(events, event)
	})

	return events
}

func (c *Crawler) CreateEvent(time time.Time, body *goquery.Selection) (*store.Event, error) {

	//attempt to parse venue address (not always set)
	venueEl := body.Find("b").First()
	addressEl := venueEl.Find("a").First()
	venueAddress := ""
	if addressEl.Length() != 0 {
		venueAddress, _ = addressEl.Attr("title")
	}

	bodyStr, err := body.Html()
	if err != nil {
		return nil, err
	}

	bodySections := regexp.MustCompile(`<br([/])?>`).Split(bodyStr, -1)
	if len(bodySections) < 1 {
		return nil, errors.New("no body")
	}

	titleLineEl, err := goquery.NewDocumentFromReader(bytes.NewBufferString(bodySections[0]))
	if err != nil {
		return nil, errors.New("invalid title line")
	}

	e := &store.Event{
		Date: time,
		Venue: &store.Venue{
			Name:    StripHTML(html.UnescapeString(venueEl.Text())),
			Address: StripHTML(html.UnescapeString(venueAddress)),
		},
		Type:        StripHTML(html.UnescapeString(strings.TrimSpace(strings.Split(titleLineEl.Text(), ":")[1]))),
		Description: StripHTML(html.UnescapeString(strings.Join(bodySections[1:], "\n"))),
	}

	//populate performers from description
	e.GuessPerformers()

	return e, nil
}
