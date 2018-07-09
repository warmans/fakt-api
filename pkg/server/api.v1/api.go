package api

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/warmans/fakt-api/pkg/server/api.v1/handler"
	mw "github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
	"github.com/warmans/fakt-api/pkg/server/data/store/event"
	"github.com/warmans/fakt-api/pkg/server/data/store/performer"
	"github.com/warmans/fakt-api/pkg/server/data/store/tag"
	"github.com/warmans/fakt-api/pkg/server/data/store/venue"
	"github.com/warmans/route-rest/routes"
	"go.uber.org/zap"
)

type API struct {
	AppVersion     string
	EventStore     *event.Store
	VenueStore     *venue.Store
	PerformerStore *performer.Store
	TagStore       *tag.Store

	Logger       *zap.Logger
}

func (a *API) NewServeMux() http.Handler {

	//entities
	restRouter := routes.GetRouter(
		[]*routes.Route{
			routes.NewRoute(
				"event",
				"{event_id:[0-9]+}",
				handler.NewEventHandler(a.EventStore),
				[]*routes.Route{
					routes.NewRoute(
						"similar",
						"{similar_id:[0-9]+}",
						handler.NewEventSimilarHandler(a.EventStore),
						[]*routes.Route{},
					),
				},
			),
			routes.NewRoute(
				"event_type",
				"{event_type_id:[0-9]+}",
				handler.NewEventTypeHandler(a.EventStore),
				[]*routes.Route{},
			),
			routes.NewRoute(
				"venue",
				"{venue_id:[0-9]+}",
				handler.NewVenueHandler(a.VenueStore),
				[]*routes.Route{
					routes.NewRoute(
						"event",
						"{event_id:[0-9]+}",
						handler.NewVenueEventHandler(a.EventStore, a.VenueStore),
						[]*routes.Route{},
					),
				},
			),
			routes.NewRoute(
				"performer",
				"{performer_id:[0-9]+}",
				handler.NewPerformerHandler(a.PerformerStore),
				[]*routes.Route{
					routes.NewRoute(
						"event",
						"{event_id:[0-9]+}",
						handler.NewPerformerEventHandler(a.EventStore, a.PerformerStore),
						[]*routes.Route{},
					),
					routes.NewRoute(
						"similar",
						"{similar_id:[0-9]+}",
						handler.NewPerformerSimilarHandler(a.PerformerStore),
						[]*routes.Route{},
					),
				},
			),
			routes.NewRoute(
				"tag",
				"{tag_id:[0-9]+}",
				handler.NewTagHandler(a.TagStore),
				[]*routes.Route{},
			),
		},
		[]string{""}, //no prefix on root resource
	)

	//meta
	restRouter.Handle("/version", handler.NewVersionHandler(a.AppVersion))

	//additional middlewares

	finalHandler := context.ClearHandler(restRouter)

	finalHandler = mw.AddCommonHeaders(
		finalHandler,
		map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers":     "Content-Type, *",
		},
	)

	return mw.AddSetup(mw.AddCommonCtx(finalHandler, a.Logger))
}
