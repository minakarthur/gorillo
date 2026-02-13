package router

import (
	"net/http"

	"github.com/engine/gorillo/internal/config"
	"github.com/engine/gorillo/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

func New(db *sqlx.DB, tmpl *handlers.TemplateRenderer, staticDir string, cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Static files
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	journal := handlers.NewJournalHandler(db, tmpl, cfg.Journal.RecentCount)
	logs := handlers.NewLogsHandler(db, tmpl)
	settings := handlers.NewSettingsHandler(db, tmpl)
	export := handlers.NewExportHandler(db)
	api := handlers.NewAPIHandler(db)
	contest := handlers.NewContestHandler(db, tmpl)

	// Pages
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/journal", http.StatusSeeOther)
	})
	r.Get("/journal", journal.Page)
	r.Post("/journal", journal.Create)

	r.Get("/logs", logs.Page)
	r.Delete("/logs", logs.Delete)

	r.Get("/settings", settings.Page)
	r.Put("/settings", settings.Save)

	r.Get("/export/cabrillo", export.Cabrillo)

	// Contests CRUD
	r.Post("/contests", contest.Create)
	r.Put("/contests/{id}", contest.Update)
	r.Delete("/contests/{id}", contest.Delete)

	// API
	r.Get("/api/callsigns", api.CallsignSuggest)
	r.Get("/api/contests/{id}", contest.GetJSON)

	return r
}
