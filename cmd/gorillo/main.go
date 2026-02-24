package main

import (
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/engine/gorillo/internal/config"
	"github.com/engine/gorillo/internal/database"
	"github.com/engine/gorillo/internal/handlers"
	"github.com/engine/gorillo/internal/router"
	"github.com/engine/gorillo/internal/timezone"
	"github.com/engine/gorillo/web"
	"github.com/jmoiron/sqlx"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	locFn := cachedLocationFunc(db)
	tmpl := handlers.NewTemplateRenderer(web.FS, locFn)
	if err := tmpl.Load(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	staticFS, _ := fs.Sub(web.FS, "static")
	mux := router.New(db, tmpl, staticFS, cfg)

	log.Printf("Gorillo starting on http://localhost:%s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// cachedLocationFunc returns a function that loads the timezone from DB
// with a 30-second cache to avoid hitting the DB on every request.
func cachedLocationFunc(db *sqlx.DB) func() *time.Location {
	var (
		mu      sync.Mutex
		cached  *time.Location
		expires time.Time
	)
	return func() *time.Location {
		mu.Lock()
		defer mu.Unlock()
		if cached != nil && time.Now().Before(expires) {
			return cached
		}
		var tz string
		if err := db.Get(&tz, "SELECT timezone FROM owners WHERE id = 1"); err != nil || tz == "" {
			cached = time.UTC
		} else {
			cached = timezone.Location(tz)
		}
		expires = time.Now().Add(30 * time.Second)
		return cached
	}
}
