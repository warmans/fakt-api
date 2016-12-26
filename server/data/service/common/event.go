package common

import (
	"regexp"
	"strings"
	"time"
)

type EventVisitor interface {
	Visit(event *Event)
}

type Event struct {
	ID          int64        `json:"id"`
	Date        time.Time    `json:"date"`
	Venue       *Venue       `json:"venue,omitempty"`
	Type        string       `json:"type"`
	Description string       `json:"description"`
	Performers  []*Performer `json:"performer,omitempty"`
	UTags       []UTags      `json:"utag"`
	Source      string       `json:"source"`
}

func (e *Event) GuessPerformers() {

	//reset performer list
	e.Performers = make([]*Performer, 0)

	re := regexp.MustCompile(`"[^"]+"[\s]+?\([^(]+\)`)
	spaceRe := regexp.MustCompile(`\"[\s]+\(`)
	fromRe := regexp.MustCompile(`(aus|from)\s+([^,\.\;]+)`)

	result := re.FindAllString(e.Description, -1)
	for _, raw := range result {
		parts := spaceRe.Split(raw, -1)
		if len(parts) != 2 {
			continue
		}

		name := strings.Trim(parts[0], `" `)
		genre := strings.Trim(parts[1], "() ")

		//try and find a location in the genre
		home := ""
		if fromMatch := fromRe.FindStringSubmatch(genre); len(fromMatch) == 3 {
			//e.g. from Berlin, from, Berlin
			home = fromMatch[2]
			//if a location was found remove it from the genre
			genre = strings.Replace(genre, fromMatch[0], "", -1)
		}

		tags := strings.Split(genre, ",")
		tags = append(tags, home)

		perf := &Performer{
			Name:  name,
			Genre: genre,
			Home:  home,
			Tags: tags,
		}
		e.Performers = append(e.Performers, perf)
	}
}

func (e *Event) IsValid() bool {
	if e.Date.IsZero() || e.Venue == nil {
		return false
	}
	return true
}

func (e *Event) Accept(visitor EventVisitor) {
	visitor.Visit(e)
}

func (e *Event) HasUTag(tag string, username string) bool {
	if tag == "" {
		return true
	}
	for _, utag := range e.UTags {
		if utag.HasValue(tag) {
			if username != "" && utag.Username != username {
				continue
			}
			return true
		}
	}
	return false
}
