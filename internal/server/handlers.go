package server

import (
	"homelab-dashboard/internal/handlers"
	"homelab-dashboard/internal/middlewares"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func setupRouter(ctx *middlewares.AppContext) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(ctx.SessionManager.LoadAndSave)

	r.Use(middlewares.AppContextMiddleware(ctx))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   ctx.Config.CORS.AllowedOrigins,
		AllowedMethods:   ctx.Config.CORS.AllowedMethods,
		AllowedHeaders:   ctx.Config.CORS.AllowedHeaders,
		ExposedHeaders:   ctx.Config.CORS.ExposedHeaders,
		AllowCredentials: ctx.Config.CORS.AllowCredentials,
		MaxAge:           ctx.Config.CORS.MaxAgeSeconds,
	}))

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/dist/assets"))))
	r.Handle("/favicon.ico", http.FileServer(http.Dir("web/dist")))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/dist/index.html")
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/status", ctx.HandlerFunc(handlers.GETAuthStatusHandler))
			r.Get("/login", ctx.HandlerFunc(handlers.LoginHandler))
			r.Get("/callback", ctx.HandlerFunc(handlers.CallbackHandler))
			r.Post("/logout", ctx.HandlerFunc(handlers.POSTLogoutHandler))
			r.Get("/logout", ctx.HandlerFunc(handlers.POSTLogoutHandler)) //TODO: Remove this
		})

		r.Get("/data", ctx.HandlerFunc(handlers.GetMetricsGET))

		r.Route("/v1", func(r chi.Router) {
			r.Get("/health", ctx.HandlerFunc(handlers.HandlerHealth))
		})
	})

	return r
}
