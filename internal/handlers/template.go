package handlers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type TemplateRenderer struct {
	dir       string
	templates map[string]*template.Template
	mu        sync.RWMutex
	funcMap   template.FuncMap
	locFn     func() *time.Location
}

func NewTemplateRenderer(dir string, locFn func() *time.Location) *TemplateRenderer {
	toLocal := func(d time.Time, t string) time.Time {
		h, m := 0, 0
		if len(t) >= 4 {
			fmt.Sscanf(t, "%02d%02d", &h, &m)
		}
		utc := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, time.UTC)
		return utc.In(locFn())
	}

	return &TemplateRenderer{
		dir:       dir,
		templates: make(map[string]*template.Template),
		locFn:     locFn,
		funcMap: template.FuncMap{
			"seq": func(start, end int) []int {
				s := make([]int, 0, end-start+1)
				for i := start; i <= end; i++ {
					s = append(s, i)
				}
				return s
			},
			"sub":   func(a, b int) int { return a - b },
			"add":   func(a, b int) int { return a + b },
			"eqs":   func(a, b string) bool { return a == b },
			"eqi":   func(a, b int) bool { return a == b },
			"slice": func(args ...string) []string { return args },
			"deref": func(p *int) int {
				if p == nil {
					return 0
				}
				return *p
			},
			"fmtDate": func(v any) string {
				switch d := v.(type) {
				case time.Time:
					return d.Format("2006-01-02")
				case string:
					return d
				default:
					return ""
				}
			},
			"localDate": func(d time.Time, t string) string {
				return toLocal(d, t).Format("2006-01-02")
			},
			"localTime": func(d time.Time, t string) string {
				return toLocal(d, t).Format("1504")
			},
			"tzName": func() string {
				return locFn().String()
			},
		},
	}
}

func (tr *TemplateRenderer) Load() error {
	layout := filepath.Join(tr.dir, "layout.html")
	logsTable := filepath.Join(tr.dir, "logs_table.html")

	contestTable := filepath.Join(tr.dir, "partials", "contest_table.html")

	// journal.html - standalone page
	journalT, err := template.New("").Funcs(tr.funcMap).ParseFiles(layout, filepath.Join(tr.dir, "journal.html"))
	if err != nil {
		return err
	}
	tr.templates["journal.html"] = journalT

	// settings.html includes contest_table.html sub-template
	settingsT, err := template.New("").Funcs(tr.funcMap).ParseFiles(layout, filepath.Join(tr.dir, "settings.html"), contestTable)
	if err != nil {
		return err
	}
	tr.templates["settings.html"] = settingsT

	// logs.html includes logs_table.html sub-template
	logsT, err := template.New("").Funcs(tr.funcMap).ParseFiles(layout, filepath.Join(tr.dir, "logs.html"), logsTable)
	if err != nil {
		return err
	}
	tr.templates["logs.html"] = logsT

	partialFiles := map[string]string{
		"logs_table.html":            filepath.Join(tr.dir, "logs_table.html"),
		"partials/qso_row.html":      filepath.Join(tr.dir, "partials", "qso_row.html"),
		"partials/contest_table.html": filepath.Join(tr.dir, "partials", "contest_table.html"),
	}
	for key, path := range partialFiles {
		t, err := template.New("").Funcs(tr.funcMap).ParseFiles(path)
		if err != nil {
			return err
		}
		tr.templates[key] = t
	}

	return nil
}

func (tr *TemplateRenderer) Render(w http.ResponseWriter, name string, data any) {
	tr.mu.RLock()
	t, ok := tr.templates[name]
	tr.mu.RUnlock()
	if !ok {
		http.Error(w, "template not found: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.ExecuteTemplate(w, "layout", data)
}

func (tr *TemplateRenderer) RenderPartial(w io.Writer, name string, data any) error {
	key := strings.ReplaceAll(name, "\\", "/")
	tr.mu.RLock()
	t, ok := tr.templates[key]
	tr.mu.RUnlock()
	if !ok {
		return nil
	}
	return t.ExecuteTemplate(w, filepath.Base(key), data)
}

func (tr *TemplateRenderer) Location() *time.Location {
	return tr.locFn()
}
