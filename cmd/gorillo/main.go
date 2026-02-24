package main

import (
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/engine/gorillo/internal/config"
	"github.com/engine/gorillo/internal/database"
	"github.com/engine/gorillo/internal/handlers"
	"github.com/engine/gorillo/internal/router"
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
			cached = locationFromTimezone(tz)
		}
		expires = time.Now().Add(30 * time.Second)
		return cached
	}
}

func locationFromTimezone(tz string) *time.Location {
	if min, ok := parseOffsetMinutes(tz); ok {
		if min == 0 {
			return time.UTC
		}
		return time.FixedZone(formatOffsetName(min), min*60)
	}

	loc, err := time.LoadLocation(strings.TrimSpace(tz))
	if err != nil {
		return time.UTC
	}
	return loc
}

func parseOffsetMinutes(tz string) (int, bool) {
	v := strings.TrimSpace(strings.ToUpper(tz))
	if v == "UTC" || v == "Z" {
		return 0, true
	}
	if len(v) < 6 || len(v) > 7 {
		return 0, false
	}
	sign := v[0]
	if sign != '+' && sign != '-' {
		return 0, false
	}

	parts := strings.Split(v[1:], ":")
	if len(parts) != 2 {
		return 0, false
	}
	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil || h < 0 || h > 14 || m < 0 || m > 59 {
		return 0, false
	}

	total := h*60 + m
	if sign == '-' {
		total = -total
	}
	if total < -12*60 || total > 14*60 {
		return 0, false
	}
	return total, true
}

func formatOffsetName(totalMinutes int) string {
	sign := "+"
	if totalMinutes < 0 {
		sign = "-"
		totalMinutes = -totalMinutes
	}
	h := totalMinutes / 60
	m := totalMinutes % 60
	return sign + pad2(h) + ":" + pad2(m)
}

func pad2(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}

	return strconv.Itoa(v)
}
