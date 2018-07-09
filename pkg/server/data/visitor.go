package data

import (
	"fmt"
	"time"

	"github.com/warmans/fakt-api/pkg/server/data/media"
	"github.com/warmans/fakt-api/pkg/server/data/service/common"
	"github.com/warmans/fakt-api/pkg/server/data/service/performer"
	"github.com/warmans/go-bandcamp-search/bcamp"
	"go.uber.org/zap"
)

// BandcampVisitor embellishes event with data from Bandcamp
type BandcampVisitor struct {
	Bandcamp    *bcamp.Bandcamp
	Logger      *zap.Logger
	ImageMirror *media.ImageMirror
}

func (v *BandcampVisitor) Visit(e *common.Event) {

	for k, performer := range e.Performers {

		if performer.ID > 0 || performer.ListenURL != "" {
			continue //don't re-fetch data for existing performer or performer with existing listen URL
		}
		//update listen URLs with bandcamp
		results, err := v.Bandcamp.Search(performer.Name, performer.Home, 1)
		if err != nil {
			v.Logger.Error("Failed to query bandcamp", zap.Error(err))
			return
		}
		if len(results) > 0 {

			imageName := performer.GetNameHash()
			if imageName == "" {
				//name was blank store images with some other hopefully unique enough number
				imageName = fmt.Sprintf("%d%d", k, time.Now().UnixNano())
			}
			//store various sized images locally instead of hot-linking original
			images, err := v.ImageMirror.Mirror(results[0].Art, imageName)
			if err != nil {
				v.Logger.Error("Failed to mirror artist images", zap.Error(err))
			} else {
				e.Performers[k].Images = images
			}
			performer.ListenURL = results[0].URL
			performer.Tags = results[0].Tags

			//get some more data
			artistInfo, err := v.Bandcamp.GetArtistPageInfo(results[0].URL)
			if err != nil {
				v.Logger.Error("Failed to get artist info", zap.Error(err))
				//don't return - use blank info
			}

			e.Performers[k].Info = artistInfo.Bio

			//embeddable player URL
			e.Performers[k].EmbedURL = bcamp.TransformEmbed(artistInfo.Embed, map[string]string{"size": "small", "bgcol": "ffffff", "linkcol": "333333", "artwork": "none", "transparent": "true"})

			for _, link := range artistInfo.Links {
				if e.Performers[k].Links == nil {
					performer.Links = make([]*common.Link, 0)
				}
				performer.Links = append(performer.Links, &common.Link{URI: link.URI, Text: link.Text})
			}

			v.Logger.Debug(fmt.Sprintf("Search Result: %+v", results[0]))
			v.Logger.Debug(fmt.Sprintf("Arist Info: %+v", artistInfo))
		}
	}
}

// EventStoreVisitor embellishes event with data from local event store
// this essentially just adds data we have already found in a previous
// update to the incoming record so we can avoid re-fetching stuff.
type PerformerServiceVisitor struct {
	PerformerService *performer.PerformerService
	Logger           *zap.Logger
}

func (v *PerformerServiceVisitor) Visit(e *common.Event) {
	//just replace whole performer if an existing one is found
	for k, perf := range e.Performers {
		existing, err := v.PerformerService.FindPerformers(&performer.Filter{Name: perf.Name, Genre: perf.Genre})
		if err != nil {
			v.Logger.Error("Failed to find perfomer visiting event", zap.Error(err))
			return
		}
		if len(existing) > 0 {
			e.Performers[k] = existing[0]
		}
	}
}
