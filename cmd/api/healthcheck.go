package main

import (
	"net/http"
)

type SystemInformation struct {
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

type Status struct {
	Available         string            `json:"status"`
	SystemInformation SystemInformation `json:"system_information"`
}

// @Summary      Show server information
// @Description  show server status and information
// @Tags         Server
// @Produce      json
// @Success      200  {object} Status
// @Failure      500  {object} Error
// @Router       /healthcheck [get]
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
