package entity

import (
	"github.com/warmans/stressfaktor-api/data/source/bcamp"
	"log"
)

type EventVisitor interface {
	Visit(event *Event)
}

// BandcampVisitor embellishes event with data from Bandcamp
type BandcampVisitor struct {
	Bandcamp *bcamp.Bandcamp
}

func (v *BandcampVisitor) Visit(e *Event) {

	for k, performer := range e.Performers {
		//update listen URLs with bandcamp
		results, err := v.Bandcamp.Search(performer.Name, performer.Home)
		if err != nil {
			log.Print("Failed to query bandcamp: %s", err.Error())
		}
		if err == nil && len(results) > 0 {
			log.Printf("%s is probably %s (%d)", performer.Name, results[0].Name, results[0].Score)
			if results[0].Score <= 1 {
				e.Performers[k].ListenURL = results[0].URL
			}
		}
	}
	//todo: add download and store thumbnail path
}

// EventStoreVisitor embellishes event with data from local event store
// this essentially just adds data we have already found in a previous
// update to the incoming record so we can avoid re-fetching stuff.
type EventStoreVisitor struct {
	Store *EventStore
}

func (v *EventStoreVisitor) Visit(e *Event) {
	//todo
}
