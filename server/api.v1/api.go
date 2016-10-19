package api

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/warmans/fakt-api/server/api.v1/handler"
	mw "github.com/warmans/fakt-api/server/api.v1/middleware"
	"github.com/warmans/fakt-api/server/data/service/event"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/warmans/fakt-api/server/data/service/user"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"github.com/warmans/fakt-api/server/data/service/venue"
	"github.com/warmans/resty"
)

type API struct {
	AppVersion       string
	EventService     *event.EventService
	VenueService     *venue.VenueService
	PerformerService *performer.PerformerService
	UserService      *user.UserStore
	UTagService      *utag.UTagService

	SessionStore sessions.Store
}

func (a *API) NewServeMux() http.Handler {
	mux := mux.NewRouter()

	mux.Handle(
		"/event",
		mw.AddCtx(resty.Restful(handler.NewEventHandler(a.EventService)), a.SessionStore, a.UserService, false),
	)
	mux.Handle(
		"/event/{id:[0-9]+}/tag",
		mw.AddCtx(handler.NewEventTagHandler(a.UTagService), a.SessionStore, a.UserService, false),
	)
	mux.Handle(
		"/event_type",
		mw.AddCtx(handler.NewEventTypeHandler(a.EventService), a.SessionStore, a.UserService, false),
	)
	mux.Handle(
		"/venue",
		mw.AddCtx(handler.NewVenueHandler(a.VenueService), a.SessionStore, a.UserService, false),
	)
	mux.Handle(
		"/performer",
		mw.AddCtx(handler.NewPerformerHandler(a.PerformerService), a.SessionStore, a.UserService, false),
	)
	mux.Handle(
		"/performer/{id:[0-9]+}/tag",
		mw.AddCtx(handler.NewPerformerTagHandler(a.UTagService), a.SessionStore, a.UserService, false),
	)

	//user
	mux.Handle(
		"/login",
		mw.AddCtx(handler.NewLoginHandler(a.UserService, a.SessionStore), a.SessionStore, a.UserService, false),
	)
	mux.Handle(
		"/logout",
		mw.AddCtx(handler.NewLogoutHandler(a.SessionStore), a.SessionStore, a.UserService, false),
	)

	mux.Handle(
		"/register",
		mw.AddCtx(handler.NewRegisterHandler(a.UserService, a.SessionStore), a.SessionStore, a.UserService, false),
	)
	mux.Handle(
		"/me",
		mw.AddCtx(handler.NewMeHandler(), a.SessionStore, a.UserService, true),
	)

	//meta
	mux.Handle("/version", handler.NewVersionHandler(a.AppVersion))

	//additional middlewares

	handler := context.ClearHandler(mux)

	handler = mw.AddCommonHeaders(
		handler,
		map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers":     "Content-Type, *",
		},
	)

	return mw.AddSetup(handler)
}

func handleAll(mux *mux.Router, handler http.Handler, routes ...string) {
	for _, route := range routes {
		mux.Handle(route, handler)
	}
}
