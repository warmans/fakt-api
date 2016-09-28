package api

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/warmans/fakt-api/server/api.v1/handler"
	mw "github.com/warmans/fakt-api/server/api.v1/middleware"
	"github.com/warmans/fakt-api/server/data/store"
	"github.com/warmans/resty"
)

type API struct {
	AppVersion   string
	DataStore    *store.Store
	UserStore    *store.UserStore
	SessionStore sessions.Store
}

func (a *API) NewServeMux() http.Handler {
	mux := mux.NewRouter()

	mux.Handle(
		"/event",
		mw.AddCtx(resty.Restful(handler.NewEventHandler(a.DataStore)), a.SessionStore, a.UserStore, false),
	)
	mux.Handle(
		"/event/{id:[0-9]+}/tag",
		mw.AddCtx(handler.NewEventTagHandler(a.DataStore), a.SessionStore, a.UserStore, false),
	)
	mux.Handle(
		"/event_type",
		mw.AddCtx(handler.NewEventTypeHandler(a.DataStore), a.SessionStore, a.UserStore, false),
	)
	mux.Handle(
		"/venue",
		mw.AddCtx(handler.NewVenueHandler(a.DataStore), a.SessionStore, a.UserStore, false),
	)
	mux.Handle(
		"/performer",
		mw.AddCtx(handler.NewPerformerHandler(a.DataStore), a.SessionStore, a.UserStore, false),
	)
	mux.Handle(
		"/performer/{id:[0-9]+}/tag",
		mw.AddCtx(handler.NewPerformerTagHandler(a.DataStore), a.SessionStore, a.UserStore, false),
	)

	//user
	mux.Handle(
		"/login",
		mw.AddCtx(handler.NewLoginHandler(a.UserStore, a.SessionStore), a.SessionStore, a.UserStore, false),
	)
	mux.Handle(
		"/logout",
		mw.AddCtx(handler.NewLogoutHandler(a.SessionStore), a.SessionStore, a.UserStore, false),
	)

	mux.Handle(
		"/register",
		mw.AddCtx(handler.NewRegisterHandler(a.UserStore, a.SessionStore), a.SessionStore, a.UserStore, false),
	)
	mux.Handle(
		"/me",
		mw.AddCtx(handler.NewMeHandler(), a.SessionStore, a.UserStore, true),
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

func handleAll(mux *mux.Router, handler http.Handler, routes... string) {
	for _, route := range routes {
		mux.Handle(route, handler)
	}
}