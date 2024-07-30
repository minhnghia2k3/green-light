package main

import (
	"context"
	"github.com/minhnghia2k3/greenlight/internal/data"
	"net/http"
)

// Define a custom contextKey
type contextKey string

// Convert the string "user" to a contextKey type and assign it to the userContextKey
const userContextKey = contextKey("user")

// The contextSetUser method returns a new copy of the request with the provided
// User struct added to the context. Note that we use our userContextKey constant as the key.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// The contextGetUser() retrieves the User struct from the request context.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
