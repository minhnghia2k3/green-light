package main

import (
	"errors"
	"expvar"
	"fmt"
	"github.com/minhnghia2k3/greenlight/internal/data"
	"github.com/pascaldekloe/jwt"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// ================== APPLICATION MIDDLEWARES
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer function which always be run in the event of panic
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex // mutual exclusion lock - prevent race condition
		clients = make(map[string]*client)
	)

	// Background goroutine which removes old entries from the clients map
	go func() {
		for {
			// cleans every minute
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks
			mu.Lock()
			// Loop through all clients. If they haven't been seen last three minutes
			// delete the corresponding entry from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass rate limiting for swagger routes
		if strings.HasPrefix(r.URL.Path, "/swagger") {
			next.ServeHTTP(w, r)
			return
		}

		if app.config.limiter.enabled {
			// Use the realip.FromRequest() function to get the client's real IP address.
			ip := realip.FromRequest(r)

			// Lock the mutex to prevent executed concurrently
			mu.Lock()

			// Check if the IP address already exists in the map
			if _, exist := clients[ip]; !exist {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			// Update the last seen time for the client
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			// Loop through the list of trusted origins, checking to see
			// if the request origin matches one of them
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// Set the preflight response headers
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authentication, Content-Type")

						w.WriteHeader(http.StatusOK)
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// ================== AUTHENTICATION MIDDLEWARES
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This header indicates to any caches that the response may vary based on the value of
		// the Authorization
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of Authorization header from the request
		authorizationHeader := r.Header.Get("Authorization")

		// If no header found
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		// Parse the JWT and extract the claims
		claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secret))
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Validate claims
		switch {
		case !claims.Valid(time.Now()):
			app.invalidAuthenticationTokenResponse(w, r)
			return
		case claims.Issuer != app.config.jwt.issuer:
			app.invalidAuthenticationTokenResponse(w, r)
			return
		case !claims.AcceptAudience(app.config.jwt.issuer):
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// At this point, JWT is OK. Extract the userID from the claims subject and convert into int64
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Look up the user record from the database
		user, err := app.models.Users.Get(userID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Add the user record to the request context
		r = app.contextSetUser(r, user)

		// Call the next handler chain
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user information from the request context
		user := app.contextGetUser(r)

		// If the user not activated, return 403 error
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		// Call the next handler in chain
		next.ServeHTTP(w, r)
	})

	// Wrap fn with the requireAuthenticatedUser() middleware before returning it.
	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		// Get user's permissions
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Check permissions slice contain code
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireActivatedUser(fn)
}

// ================== METRICS MIDDLEWARES
func (app *application) metrics(next http.Handler) http.Handler {
	// Initialize the new expvar variables when the middleware chain is first built
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
		totalResponseSentByStatus       = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add 1 per request
		totalRequestsReceived.Add(1)

		mw := &metricsResponseWriter{wrapped: w}

		next.ServeHTTP(mw, r)

		// Add 1 per response
		totalResponsesSent.Add(1)

		totalResponseSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)

		// Calculate the number of microseconds since we began to process the request,
		// then increment the total processing time by this amount.
		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}

func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !mw.headerWritten {
		mw.statusCode = http.StatusOK
		mw.headerWritten = true
	}

	return mw.wrapped.Write(b)
}

func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}
