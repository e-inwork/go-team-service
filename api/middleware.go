// Copyright 2022, e-inwork.com. All rights reserved.

package api

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/e-inwork-com/go-team-service/pkg/data"
	"github.com/google/uuid"

	"github.com/felixge/httpsnoop"
	"github.com/golang-jwt/jwt/v4"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// Claims JSON Web Token
type Claims struct {
	ID uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *Application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Config.Limiter.Enabled {
			ip := realip.FromRequest(r)

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.Config.Limiter.Rps), app.Config.Limiter.Burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.Config.Cors.TrustedOrigins {
				if origin == app.Config.Cors.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {

						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("profile_total_requests_received")
	totalResponsesSent := expvar.NewInt("profile_total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("profile_total_processing_time_μs")

	totalResponsesSentByStatus := expvar.NewMap("profile_total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		totalRequestsReceived.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		totalResponsesSent.Add(1)

		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}

func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get an Authorization header from HTTP request
		// and if not available then just return to next function
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Split a header and get a Bearer part
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get a token
		tokenString := headerParts[1]

		//  Define clain from the API user
		claims := &Claims{}

		// Parse a token with the secret
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(app.Config.Auth.Secret), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				app.invalidCredentialsResponse(w, r)
				return
			}
			app.badRequestResponse(w, r, err)
			return
		}

		//  Check if the token is valid
		if !token.Valid {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get a user by ID from the Claim token
		user, err := app.Models.Users.GetByID(claims.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Put the user inside a context to use it on the next function
		r = app.contextSetUser(r, user)

		// Run the next function
		next.ServeHTTP(w, r)
	})
}

func (app *Application) requireAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		// If the user doesn't have an authentication then send an error
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		// Run the next function
		next.ServeHTTP(w, r)
	})
}
