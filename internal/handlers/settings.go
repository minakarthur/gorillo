package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/engine/gorillo/internal/i18n"
	"github.com/engine/gorillo/internal/models"
	"github.com/engine/gorillo/internal/timezone"
	"github.com/jmoiron/sqlx"
)

type SettingsHandler struct {
	db   *sqlx.DB
	tmpl *TemplateRenderer
}

func NewSettingsHandler(db *sqlx.DB, tmpl *TemplateRenderer) *SettingsHandler {
	return &SettingsHandler{db: db, tmpl: tmpl}
}

func (h *SettingsHandler) Page(w http.ResponseWriter, r *http.Request) {
	var owner models.Owner
	h.db.Get(&owner, "SELECT id, callsign, name, email, grid_locator, location, club, address, operators, timezone, language FROM owners WHERE id = 1")

	var contests []models.Contest
	h.db.Select(&contests, "SELECT * FROM contests ORDER BY name")

	lang := i18n.ParseLang(owner.Language)
	selectedTZ := normalizeTimezoneValue(owner.Timezone)

	h.tmpl.Render(w, "settings.html", map[string]any{
		"Owner":            owner,
		"Active":           "settings",
		"Contests":         contests,
		"Modes":            modeList,
		"Freqs":            freqList,
		"Timezones":        timezoneList,
		"SelectedTimezone": selectedTZ,
		"T":                i18n.TFunc(lang),
		"Lang":             string(lang),
	})
}

func (h *SettingsHandler) Save(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	_, err := h.db.Exec(`UPDATE owners SET
		callsign=?, name=?, email=?, grid_locator=?, location=?, club=?,
		address=?, operators=?, timezone=?, language=?
		WHERE id = 1`,
		r.FormValue("callsign"),
		r.FormValue("name"),
		r.FormValue("email"),
		r.FormValue("grid_locator"),
		r.FormValue("location"),
		r.FormValue("club"),
		r.FormValue("address"),
		r.FormValue("operators"),
		normalizeTimezoneValue(r.FormValue("timezone")),
		r.FormValue("language"),
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		lang := i18n.ParseLang(r.FormValue("language"))
		t := i18n.TFunc(lang)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div id="save-status" class="text-green-600 text-sm font-medium">` + t("settings.saved") + `</div>`))
		return
	}
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

var timezoneList = []string{
}

func init() {
	timezoneList = buildTimezoneList()
}

func buildTimezoneList() []string {
	tzs := []string{"UTC"}
	for min := -12 * 60; min <= 14*60; min += 15 {
		if min == 0 {
			continue
		}
		tzs = append(tzs, timezone.FormatOffset(min))
	}
	return tzs
}

func normalizeTimezoneValue(tz string) string {
	if min, ok := timezone.ParseOffsetMinutes(tz); ok {
		if min == 0 {
			return "UTC"
		}
		return timezone.FormatOffset(min)
	}

	loc, err := time.LoadLocation(strings.TrimSpace(tz))
	if err == nil {
		_, offSec := time.Now().In(loc).Zone()
		min := offSec / 60
		if min == 0 {
			return "UTC"
		}
		return timezone.FormatOffset(min)
	}

	return "UTC"
}
