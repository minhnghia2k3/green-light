package main

import (
	"database/sql"
	"errors"
	"github.com/minhnghia2k3/greenlight/internal/data"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"net/http"
	"time"
)

type UserResponse struct {
	User data.User `json:"user"`
}

type RegisterUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenInput struct {
	TokenPlainText string `json:"token"`
}

type UpdateUserInput struct {
	Password       string `json:"password"`
	TokenPlainText string `json:"token"`
}

// registerUserHandler function handle register a new user, and sending email in the background.
// @Summary      Register account
// @Description  register user account
// @Param input body RegisterUserInput true "register user input"
// @Tags         Users
// @Accept 		 json
// @Produce      json
// @Success      201  {object} UserResponse
// @Failure      400  {object} Error
// @Failure      422  {object} Error
// @Failure      500  {object} Error
// @Router       /users [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input RegisterUserInput

	// Parse the request body
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	// Copy input into User struct
	user := data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Generate and store the hashed and plaintext passwords
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Validate input
	v := validation.New()
	if data.ValidateUser(v, &user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert new user record into database
	err = app.models.Users.Insert(&user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Activation token
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Background goroutine to send email
	app.background(func() {

		dynamicData := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		// Send welcome email
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", dynamicData)
		if err != nil {
			// If error, use log instead sending http error.
			app.logger.PrintError(err, nil)
			return
		}
		app.logger.PrintInfo("sending email successfully", nil)
	})

	// Add `movie:read` default permission
	err = app.models.Permissions.AddForUser(user.ID, "movie:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// http.StatusAccepted indicates that the request has been accepted,
	// but processing has not been completed.
	err = app.writeJSON(w, http.StatusAccepted, envelop{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// @Summary      Activate user account
// @Description  activate user account
// @Param   input      body UpdateUserInput true "update user input"
// @Tags         Users
// @Accept 		 json
// @Produce      json
// @Success      200  {object} UserResponse
// @Failure      400  {object} Error
// @Failure      401  {object} Error
// @Failure      422  {object} Error
// @Failure      500  {object} Error
// @Router       /users/activated [put]
func (app *application) activeUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse input data
	var input TokenInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	// Validation check
	v := validation.New()

	if data.ValidateTokenPlainText(v, input.TokenPlainText); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Validate validity of provided token
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update user activation
	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Delete correspond token if everything went successfully
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the updated user details to the client
	err = app.writeJSON(w, http.StatusOK, envelop{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Verify the password reset token and set a new password for the user
// @Summary      Change user current password
// @Description  change user current password, must provide valid *reset password* token.
// @Param   input      body TokenInput true "plain text token input"
// @Tags         Users
// @Accept 		 json
// @Produce      json
// @Success      200  {object} UserResponse
// @Failure      400  {object} Error
// @Failure      422  {object} Error
// @Failure      500  {object} Error
// @Router       /users/password [put]
func (app *application) updateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the user's new password and password reset token.
	var input UpdateUserInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	v := validation.New()

	data.ValidatePasswordPlaintext(v, input.Password)
	data.ValidateTokenPlainText(v, input.TokenPlainText)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Check token plain text in the database
	user, err := app.models.Users.GetForToken(data.ScopePasswordReset, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired password reset token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Hash plain password and store into pointer `user` struct
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrConflictEdit):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Delete all password reset tokens for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopePasswordReset, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send user confirm message
	env := envelop{"message": "your password was successfully reset"}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
