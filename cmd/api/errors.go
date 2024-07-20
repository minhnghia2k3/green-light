package main

import (
	"fmt"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"net/http"
)

// The logError() method is a generic helper for logging an error message.
func (app *application) logError(r *http.Request, err error) {
	app.logger.Print(err)
}

// The errorResponse() method is a generic helper for sending JSON-formatted
// error messages to the client with a given status code.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelop{"error": message}

	// Write the response
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// The serverErrorResponse() method will be used when our application encounters
// an unexpected problem at runtime. It logs the detailed error message, then uses
// errorResponse() helper to send a 500 INTERNAL SERVER ERROR and JSON response to the client.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.errorResponse(w, r, http.StatusBadRequest, message)

}

// The notFoundResponse() method will be used to send a 404 NOT FOUND status code and JSON response.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource cant not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// The methodNotAllow() method will be used to send a 405 Method Not Allowed status code and JSON response.
func (app *application) methodNotAllow(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not allow for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors validation.MapErrors) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
