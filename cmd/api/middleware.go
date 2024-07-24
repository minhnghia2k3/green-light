package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

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
			defer mu.Unlock()
			// Loop through all clients. If they haven't been seen last three minutes
			// delete the corresponding entry from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// Extract the client's IP
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// Lock the mutex to prevent executed concurrently
			mu.Lock()
			defer mu.Unlock()

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
		}
		next.ServeHTTP(w, r)
	})
}
