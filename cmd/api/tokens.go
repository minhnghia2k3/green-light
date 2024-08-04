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

type ResetPasswordInput struct {
	Email string `json:"email" example:"john@example.com"`
}

type TokenResponse struct {
	AuthenticationToken string `json:"authentication_token"`
}

type ResetTokenResponse struct {
	Message string `json:"message" example:"an email will be sent to you containing password reset instructions"`
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

// Generate a password reset token and send it to the user's email address
// @Summary      Request a token for reset password
// @Description  request a token for reset password, must provide *valid* and *activated* email address
// @Param   input      body ResetPasswordInput true "Reset password input"
// @Tags         Authentications
// @Accept 		 json
// @Produce      json
// @Success      202  {object} ResetTokenResponse
// @Failure      400  {object} Error
// @Failure      422  {object} Error
// @Failure      500  {object} Error
// @Router       /tokens/password-reset [post]
func (app *application) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the user's email address
	var input ResetPasswordInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	v := validation.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Find corresponding user record by given email
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if !user.Activated {
		v.AddError("email", "user account must be activated, please check your email again.")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// If checking successfully, create a new password reset token with a 45-minute expiry time
	token, err := app.models.Tokens.New(user.ID, 45*time.Minute, data.ScopePasswordReset)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Email the user with their password reset token.
	go func() {
		dynamicData := map[string]any{
			"passwordResetToken": token.Plaintext,
		}

		err = app.mailer.Send(user.Email, "token_password_reset.tmpl", dynamicData)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}

		app.logger.PrintInfo("sending reset password email successfully", nil)
	}()

	env := envelop{"message": "an email will be sent to you containing password reset instructions"}

	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
