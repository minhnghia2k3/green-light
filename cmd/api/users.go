package main

import (
	"errors"
	"github.com/minhnghia2k3/greenlight/internal/data"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"net/http"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string
		Email    string
		Password string
	}

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

	// Background goroutine to send email
	app.background(func() {
		// Send welcome email
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			// If error, use log instead sending http error.
			app.logger.PrintError(err, nil)
		}
		app.logger.PrintInfo("sending email successfully", nil)
	})

	// http.StatusAccepted indicates that the request has been accepted,
	// but processing has not been completed.
	err = app.writeJSON(w, http.StatusAccepted, envelop{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
