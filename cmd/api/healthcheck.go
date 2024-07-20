package main

import (
	"net/http"
)

// healthcheckHandler handler which writes a plain-text response
// with information about the application status, operating environment and version.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelop{
		"status": "available",
		"system_information": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	// Encode data to JSON and write out
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
