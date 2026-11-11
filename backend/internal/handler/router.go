package handler

import (
	"net/http"

	"autodidact-backend/internal/config"
	"autodidact-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg *config.Config, svc *service.SyllabusService, focusSvc *service.FocusService) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Auth-Token"},
		ExposedHeaders:   []string{"Link", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authH := NewAuthHandler(cfg.AuthToken)
	syllabusH := NewSyllabusHandler(svc)
	moduleH := NewModuleHandler(svc)
	stepH := NewStepHandler(svc)
	focusH := NewFocusHandler(focusSvc)

	// Public routes
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"app":    "autodidact",
		})
	})
	r.Post("/api/auth/verify", authH.Verify)

	// Protected routes
	r.Group(func(protected chi.Router) {
		protected.Use(authH.Middleware)

		// Focus / What Next
		protected.Get("/api/focus", focusH.List)
		protected.Post("/api/focus", focusH.Create)
		protected.Put("/api/focus/reorder", focusH.Reorder)
		protected.Put("/api/focus/{id}", focusH.Update)
		protected.Delete("/api/focus/{id}", focusH.Delete)
		protected.Get("/api/focus/backlog-items", focusH.ListBacklogItems)

		// Syllabi
		protected.Get("/api/syllabi", syllabusH.List)
		protected.Post("/api/syllabi", syllabusH.Create)
		protected.Get("/api/syllabi/{id}", syllabusH.GetByID)
		protected.Put("/api/syllabi/{id}", syllabusH.Update)
		protected.Delete("/api/syllabi/{id}", syllabusH.Delete)
		protected.Post("/api/syllabi/{id}/duplicate", syllabusH.Duplicate)
		protected.Get("/api/syllabi/{id}/active-step", syllabusH.GetActiveStep)

		// Modules
		protected.Post("/api/syllabi/{syllabusId}/modules", moduleH.Create)
		protected.Put("/api/modules/{id}", moduleH.Update)
		protected.Delete("/api/modules/{id}", moduleH.Delete)

		// Steps
		protected.Post("/api/modules/{moduleId}/steps", stepH.Create)
		protected.Get("/api/steps/{id}", stepH.GetByID)
		protected.Put("/api/steps/{id}", stepH.Update)
		protected.Delete("/api/steps/{id}", stepH.Delete)
		protected.Put("/api/steps/{id}/progress", stepH.UpdateProgress)
	})

	return r
}
