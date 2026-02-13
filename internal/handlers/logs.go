package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/engine/gorillo/internal/i18n"
	"github.com/engine/gorillo/internal/models"
	"github.com/jmoiron/sqlx"
)

const pageSize = 25

type LogsHandler struct {
	db   *sqlx.DB
	tmpl *TemplateRenderer
}

func NewLogsHandler(db *sqlx.DB, tmpl *TemplateRenderer) *LogsHandler {
	return &LogsHandler{db: db, tmpl: tmpl}
}

func (h *LogsHandler) Page(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	contestID, _ := strconv.Atoi(r.URL.Query().Get("contest_id"))

	var total int
	if contestID > 0 {
		h.db.Get(&total, "SELECT COUNT(*) FROM qso_logs WHERE contest_id = ?", contestID)
	} else {
		h.db.Get(&total, "SELECT COUNT(*) FROM qso_logs")
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	offset := (page - 1) * pageSize
	var qsos []models.QSO
	if contestID > 0 {
		h.db.Select(&qsos, "SELECT * FROM qso_logs WHERE contest_id = ? ORDER BY date DESC, time DESC LIMIT ? OFFSET ?", contestID, pageSize, offset)
	} else {
		h.db.Select(&qsos, "SELECT * FROM qso_logs ORDER BY date DESC, time DESC LIMIT ? OFFSET ?", pageSize, offset)
	}

	var contests []models.Contest
	h.db.Select(&contests, "SELECT * FROM contests ORDER BY name")

	var owner models.Owner
	h.db.Get(&owner, "SELECT language FROM owners WHERE id = 1")
	lang := i18n.ParseLang(owner.Language)

	data := map[string]any{
		"QSOs":       qsos,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
		"Active":     "logs",
		"ContestID":  contestID,
		"Contests":   contests,
		"T":          i18n.TFunc(lang),
		"Lang":       string(lang),
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		h.tmpl.RenderPartial(w, "logs_table.html", data)
		return
	}
	h.tmpl.Render(w, "logs.html", data)
}

func (h *LogsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	h.db.Exec("DELETE FROM qso_logs WHERE id = ?", id)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(""))
		return
	}
	http.Redirect(w, r, "/logs", http.StatusSeeOther)
}
