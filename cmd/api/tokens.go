package main

import (
	"errors"
	"github.com/minhnghia2k3/greenlight/internal/data"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"net/http"
	"time"
)

type LoginInput struct {
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"myP4SSw3rd"`
}

type TokenResponse struct {
	AuthenticationToken string `json:"authentication_token"`
}

// @Summary      Create authentication token
// @Description  login account by email and password
// @Tags         Authentications
// @Accept 		 json
// @Produce      json
// @Param		 input 	body 	LoginInput	true	"Login parameters"
// @Success      201  {object} TokenResponse
// @Failure      400  {object} Error
// @Failure      401  {object} Error
// @Failure      422  {object} Error
// @Failure      500  {object} Error
// @Router       /tokens/authentication [post]
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the email and password from the request body.
	var input LoginInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	// Validate the mail and password provided by the client
	v := validation.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Look up the user record based on the input
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):

			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Check if the provided password matches the actual password for the user.
	// If no matching then return app.invalidCredentialsResponse()
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// If the password don't match, call the app.invalidCredentialsResponse() helper
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Otherwise, if the password is correct, we generate a new token with a 24h expiry time
	// with the scope "authentication".
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Encode the token to JSON and send it in response along with 201 Created status code
	err = app.writeJSON(w, http.StatusCreated, envelop{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
