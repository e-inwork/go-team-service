// Copyright 2022, e-inwork.com. All rights reserved.

package api

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/service/teams/health", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/service/teams", app.requireAuthenticated(app.createTeamHandler))
	router.HandlerFunc(http.MethodGet, "/service/teams/me", app.requireAuthenticated(app.getOwnTeamHandler))
	router.HandlerFunc(http.MethodPatch, "/service/teams/:id", app.requireAuthenticated(app.patchTeamHandler))
	router.HandlerFunc(http.MethodGet, "/service/teams/pictures/:file", app.getProfilePictureHandler)
	router.HandlerFunc(http.MethodPost, "/service/teams/members", app.requireAuthenticated(app.createTeamMemberHandler))
	router.HandlerFunc(http.MethodGet, "/service/teams/members", app.requireAuthenticated(app.listTeamMembersByOwnerHandler))
	router.HandlerFunc(http.MethodDelete, "/service/teams/members/:id", app.requireAuthenticated(app.deleteTeamMemberHandler))
	router.HandlerFunc(http.MethodGet, "/service/teams/members/:id", app.requireAuthenticated(app.getTeamMemberHandler))

	router.Handler(http.MethodGet, "/service/teams/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
