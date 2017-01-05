package api

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/warmans/fakt-api/server/api.v1/handler"
	mw "github.com/warmans/fakt-api/server/api.v1/middleware"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/fakt-api/server/data/service/tag"
	"github.com/warmans/fakt-api/server/data/service/user"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"github.com/warmans/fakt-api/server/data/service/venue"
	"github.com/warmans/route-rest/routes"
)

type API struct {
	AppVersion       string
	EventService     *event.EventService
	VenueService     *venue.VenueService
	PerformerService *performer.PerformerService
	UserService      *user.UserStore
	UTagService      *utag.UTagService
	TagService       *tag.TagService

	SessionStore sessions.Store
	Logger       log.Logger
}

func (a *API) NewServeMux() http.Handler {

	//entities
	restRouter := routes.GetRouter(
		[]*routes.Route{
			routes.NewRoute(
				"event",
				"{event_id:[0-9]+}",
				handler.NewEventHandler(a.EventService),
				[]*routes.Route{
					routes.NewRoute(
						"utag",
						"{utag_id:[0-9]+}",
						handler.NewEventUTagHandler(a.UTagService),
						[]*routes.Route{},
					),
					routes.NewRoute(
						"similar",
						"{similar_id:[0-9]+}",
						handler.NewEventSimilarHandler(a.EventService),
						[]*routes.Route{},
					),
				},
			),
			routes.NewRoute(
				"event_type",
				"{event_type_id:[0-9]+}",
				handler.NewEventTypeHandler(a.EventService),
				[]*routes.Route{},
			),
			routes.NewRoute(
				"venue",
				"{venue_id:[0-9]+}",
				handler.NewVenueHandler(a.VenueService),
				[]*routes.Route{
					routes.NewRoute(
						"event",
						"{event_id:[0-9]+}",
						handler.NewVenueEventHandler(a.EventService, a.VenueService),
						[]*routes.Route{},
					),
				},
			),
			routes.NewRoute(
				"performer",
				"{performer_id:[0-9]+}",
				handler.NewPerformerHandler(a.PerformerService),
				[]*routes.Route{
					routes.NewRoute(
						"utag",
						"{utag_id:[0-9]+}",
						handler.NewPerformerUTagHandler(a.UTagService),
						[]*routes.Route{},
					),
					routes.NewRoute(
						"event",
						"{event_id:[0-9]+}",
						handler.NewPerformerEventHandler(a.EventService, a.PerformerService),
						[]*routes.Route{},
					),
					routes.NewRoute(
						"similar",
						"{similar_id:[0-9]+}",
						handler.NewPerformerSimilarHandler(a.PerformerService),
						[]*routes.Route{},
					),
				},
			),
			routes.NewRoute(
				"tag",
				"{tag_id:[0-9]+}",
				handler.NewTagHandler(a.TagService),
				[]*routes.Route{},
			),
		},
		[]string{""}, //no prefix on root resource
	)

	//user
	restRouter.Handle(
		"/login",
		handler.NewLoginHandler(a.UserService, a.SessionStore),
	)
	restRouter.Handle(
		"/logout",
		handler.NewLogoutHandler(a.SessionStore),
	)
	restRouter.Handle(
		"/register",
		handler.NewRegisterHandler(a.UserService, a.SessionStore),
	)
	restRouter.Handle(
		"/me",
		handler.NewMeHandler(),
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

	return mw.AddSetup(mw.AddCommonCtx(finalHandler, a.SessionStore, a.UserService, a.Logger))
}
