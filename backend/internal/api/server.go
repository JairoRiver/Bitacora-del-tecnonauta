package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/auth"
	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/web"
)

type Server struct {
	queries         db.Querier
	authSecret      string
	sessionDuration time.Duration
}

func NewServer(queries db.Querier, authSecret string, sessionDuration time.Duration) *Server {
	return &Server{
		queries:         queries,
		authSecret:      authSecret,
		sessionDuration: sessionDuration,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(zerologMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", s.handleHealth)

	// Public API consumed by Astro at build time
	r.Route("/api", func(r chi.Router) {
		r.Get("/posts", s.handleListPosts)
		r.Get("/posts/{slug}", s.handleGetPost)
		r.Get("/categories", s.handleListCategories)
		r.Get("/categories/{slug}", s.handleListPostsByCategory)
	})

	// Admin UI (HTML)
	webHandler := web.NewHandler(s.queries, s.authSecret, s.sessionDuration)
	r.Route("/admin", func(r chi.Router) {
		webHandler.MountPublic(r)
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(s.authSecret))
			webHandler.MountProtected(r)
		})
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func zerologMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		next.ServeHTTP(ww, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Dur("duration", time.Since(start)).
			Str("request_id", middleware.GetReqID(r.Context())).
			Msg("request")
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
