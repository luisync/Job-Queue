package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luisync/Job-Queue/internal/env"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	"github.com/luisync/Job-Queue/internal/jobs"
	"github.com/luisync/Job-Queue/internal/users"
)

// Defines how requests are handled.
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Midllewares.
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Time out a request after 60 seconds.
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes for health checks.
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ready"))
	})
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Create handlers, services, and repository.
	jobsHandler := jobs.NewHandler(app.db)

	const minSecretKeySize = 32
	var secretKey = env.GetString("SECRET_KEY", "01234567890123456789012345678901")
	if len([]rune(secretKey)) < minSecretKeySize {
		log.Fatalf("Secret key must be at least %d characters.", minSecretKeySize)
	}
	usersHandler := users.NewHandler(app.db, secretKey)
	tokenMaker := usersHandler.TokenMaker

	// Routes for jobs
	r.Group(func(r chi.Router) {
		r.Route("/jobs", func(r chi.Router) {
			r.Use(users.GetAuthMiddlewareFunc(tokenMaker))

			// Find all jobs belonging to a certain user.
			r.Get("/", jobsHandler.ListJobs)

			// Create a new job.
			r.Post("/", jobsHandler.CreateJob)

			// Route for specific jobs.
			r.Route("/{jobID}", func(r chi.Router) {
				// Find a jobs by id.
				r.Get("/", jobsHandler.FindJob)

				// Route for job results.
				r.Route("/results", func(r chi.Router) {
					// Find all job results for this job.
					r.Get("/", jobsHandler.ListJobResults)

					// Find the latest job result.
					r.Get("/latest", jobsHandler.FindLatestJobResult)
				})
			})
		})
	})

	// Routes for accounts.
	r.Route("/account", func(r chi.Router) {
		// Create a new account.
		r.Post("/register", usersHandler.Register)

		// Login to an account.
		r.Post("/login", usersHandler.Login)

		// Logout of an account.
		r.Group(func(r chi.Router) {
			r.Use(users.GetAuthMiddlewareFunc(tokenMaker))
			r.Get("/logout", usersHandler.Logout)
		})

		// Route for account tokens.
		r.Group(func(r chi.Router) {
			r.Use(users.GetAuthMiddlewareFunc(tokenMaker))
			r.Route("/tokens", func(r chi.Router) {
				// Renew an access token.
				r.Post("/renew", usersHandler.RenewAccessToken)

				// Revoke a session.
				r.Post("/revoke", usersHandler.RevokeSession)
			})
		})
	})

	return r
}

// Creates a listener for network activity.
func (app *application) run(h http.Handler) error {
	// Define the address, handler, and timeout constraints.
	srv := &http.Server{
		Addr:         app.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	// Log the server starting.
	log.Printf("Server has started at address: %s", app.addr)

	// Start listener.
	return srv.ListenAndServe()
}

// Application instance.
type application struct {
	addr string
	db   pb.JobqueueClient
}
