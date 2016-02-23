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
		if performer.ID > 0 || performer.ListenURL != "" {
			continue //don't re-fetch data for existing performer or performer with existing listen URL
		}
		//update listen URLs with bandcamp
		results, err := v.Bandcamp.Search(performer.Name, performer.Home)
		if err != nil {
			log.Print("Failed to query bandcamp: %s", err.Error())
		}
		if err == nil && len(results) > 0 {
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
	//just replace whole performer if an existing one is found
	for k, performer := range e.Performers {
		existing, err := v.Store.FindPerformers(&PerformerFilter{Name: performer.Name, Genre: performer.Genre})
		if err != nil {
			log.Print("failed to find perfomer visiting event: %s", err.Error())
			return
		}
		if len(existing) > 0 {
			e.Performers[k] = existing[0]
		}
	}
}
