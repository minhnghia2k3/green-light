package main

import (
	"expvar"
	"github.com/julienschmidt/httprouter"
	"github.com/swaggo/http-swagger"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllow)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movie:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movie:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movie:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movie:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movie:write", app.deleteMovieHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activeUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	// Authentication
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	// Metric
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Documenting
	router.Handler(http.MethodGet, "/swagger/*docs", httpSwagger.WrapHandler)

	// Wrap the router with the panic recovery middleware
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
