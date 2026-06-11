package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/milann/taskflow/internal/config"
	"github.com/milann/taskflow/internal/realtime"
	"github.com/milann/taskflow/internal/store"
)

// Server wires configuration, persistence and realtime into HTTP handlers.
type Server struct {
	cfg    *config.Config
	store  store.Store
	hub    *realtime.Hub
	logger *slog.Logger
}

func NewServer(cfg *config.Config, st store.Store, hub *realtime.Hub, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, store: st, hub: hub, logger: logger}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", s.handleSignup)
			r.Post("/login", s.handleLogin)
			r.Group(func(r chi.Router) {
				r.Use(s.requireAuth)
				r.Get("/me", s.handleMe)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)

			r.Route("/tasks", func(r chi.Router) {
				r.Post("/", s.handleCreateTask)
				r.Get("/", s.handleListTasks)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.handleGetTask)
					r.Patch("/", s.handleUpdateTask)
					r.Delete("/", s.handleDeleteTask)
					r.Get("/activity", s.handleListActivity)
					r.Post("/attachments", s.handleUploadAttachment)
					r.Get("/attachments", s.handleListAttachments)
				})
			})

			r.Route("/attachments/{id}", func(r chi.Router) {
				r.Get("/download", s.handleDownloadAttachment)
				r.Delete("/", s.handleDeleteAttachment)
			})

			r.Get("/events", s.handleEvents)
		})
	})

	return r
}
