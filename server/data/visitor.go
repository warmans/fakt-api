package data

import (
	"log"

	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/go-bandcamp-search/bcamp"
)

// BandcampVisitor embellishes event with data from Bandcamp
type BandcampVisitor struct {
	Bandcamp       *bcamp.Bandcamp
	VerboseLogging bool
}

func (v *BandcampVisitor) Visit(e *common.Event) {

	for k, performer := range e.Performers {
		if performer.ID > 0 || performer.ListenURL != "" {
			continue //don't re-fetch data for existing performer or performer with existing listen URL
		}
		//update listen URLs with bandcamp
		results, err := v.Bandcamp.Search(performer.Name, performer.Home, 1)
		if err != nil {
			log.Print("Failed to query bandcamp: %s", err.Error())
			return
		}
		if len(results) > 0 {
			//todo: IMPORTANT mirror images locally rather than hotlinking
			e.Performers[k].ListenURL = results[0].URL
			e.Performers[k].Img = results[0].Art
			e.Performers[k].Tags = results[0].Tags

			//get some more data
			artistInfo, err := v.Bandcamp.GetArtistPageInfo(results[0].URL)
			if err != nil {
				log.Print("Failed to get artist info: %s", err.Error())
				//don't return - use blank info
			}
			e.Performers[k].Info = artistInfo.Bio
			for _, link := range artistInfo.Links {
				if e.Performers[k].Links == nil {
					e.Performers[k].Links = make([]*common.Link, 0)
				}
				e.Performers[k].Links = append(e.Performers[k].Links, &common.Link{URI: link.URI, Text: link.Text})
			}
			if v.VerboseLogging {
				log.Printf("Search Result: %+v", results[0])
				log.Printf("Arist Info: %+v", artistInfo)
			}
		}
	}
}

// EventStoreVisitor embellishes event with data from local event store
// this essentially just adds data we have already found in a previous
// update to the incoming record so we can avoid re-fetching stuff.
type PerformerServiceVisitor struct {
	PerformerService *performer.PerformerService
}

func (v *PerformerServiceVisitor) Visit(e *common.Event) {
	//just replace whole performer if an existing one is found
	for k, perf := range e.Performers {
		existing, err := v.PerformerService.FindPerformers(&performer.PerformerFilter{Name: perf.Name, Genre: perf.Genre})
		if err != nil {
			log.Print("failed to find perfomer visiting event: %s", err.Error())
			return
		}
		if len(existing) > 0 {
			e.Performers[k] = existing[0]
		}
	}
}
