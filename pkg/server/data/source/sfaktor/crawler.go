package sfaktor

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-kit/kit/log"
	"github.com/warmans/fakt-api/pkg/server/data/service/common"
)

var validDate = regexp.MustCompile(`^[A-Za-z]+, [0-9]{2}\. [\p{L}]+ [0-9]{4}$`)
var validTime = regexp.MustCompile(`^[0-9]{2}:[0-9]{2}$`)

type Crawler struct {
	TermineURI string
	Timezone   *time.Location
	Logger     log.Logger
}

func (c *Crawler) Name() string {
	return "stressfaktor"
}

func (c *Crawler) Crawl() ([]*common.Event, error) {

	events := make([]*common.Event, 0)

	res, err := http.Get(c.TermineURI)
	if err != nil {
		return events, fmt.Errorf("SF crawler fetch failed: %s", err.Error())
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return events, fmt.Errorf("SF crawler failed to parse result: %s", err.Error())
	}

	//select the main data column and handle all the sub-tables
	doc.Find("#column_center").Each(func(i int, sel *goquery.Selection) {
		events = append(events, c.HandleDateTable(i, sel, c.Timezone)...)
	})

	return events, nil
}

func (c *Crawler) HandleDateTable(i int, sel *goquery.Selection, localTime *time.Location) []*common.Event {

	var dateStr string
	var dateTime time.Time
	var failed bool
	var events []*common.Event

	sel.Find(".termin_tag_titel, .termin_box").Each(func(i int, row *goquery.Selection) {

		if failed {
			return
		}

		//first row is always the date
		if dateStr == "" || row.HasClass("termin_tag_titel") {
			dateStr = strings.TrimSpace(row.Text())
			if !validDate.MatchString(dateStr) {
				c.Logger.Log("msg", fmt.Sprintf("Invalid date string: %s", dateStr))
				failed = true
			}
			return //move on
		}

		//each row has a time
		timeStr := strings.TrimSpace(row.Find(".spalte_uhrzeit > .uhrzeit2").First().Text())
		if !validTime.MatchString(timeStr) {
			c.Logger.Log("msg", fmt.Sprintf("Invalid time string: %s (text: %s)", timeStr, row.Text()))
			return //don't fail
		}

		var err error
		dateTime, err = ParseTime(dateStr, timeStr, localTime)
		if err != nil {
			c.Logger.Log("msg", fmt.Sprintf("Failed to parse date: %s %s (%s)", dateStr, timeStr, err.Error()))
			return
		}

		event, err := c.CreateEvent(dateTime, row)
		if err != nil {
			c.Logger.Log("msg", fmt.Sprintf("Failed to create event: %s", err.Error()))
			return
		}

		events = append(events, event)
	})

	return events
}

func (c *Crawler) CreateEvent(time time.Time, terminBox *goquery.Selection) (*common.Event, error) {

	//attempt to parse venue address (not always set)
	venueEl := terminBox.Find(".spalte_termintext > b").First()
	addressEl := venueEl.Find("a").First()
	venueAddress := ""
	if addressEl.Length() != 0 {
		venueAddress, _ = addressEl.Attr("title")
	}

	bodyText := terminBox.Find(".spalte_termintext")
	bodyStr, err := bodyText.Html()
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

	tags := []string{}
	terminBox.Find(".termin_tags > .termin_tag").Each(func(i int, tag *goquery.Selection) {
		tags = append(tags, tag.Text())
	})

	e := &common.Event{
		Date: time,
		Venue: &common.Venue{
			Name:    StripHTML(html.UnescapeString(venueEl.Text())),
			Address: StripHTML(html.UnescapeString(venueAddress)),
		},
		Type:        StripHTML(html.UnescapeString(strings.TrimSpace(strings.Split(titleLineEl.Text(), ":")[1]))),
		Description: strings.TrimSpace(StripHTML(html.UnescapeString(strings.Join(bodySections[1:], "\n")))),
		Tags:        tags,
	}

	//populate performers from description
	e.GuessPerformers()

	return e, nil
}
