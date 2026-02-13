package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/engine/gorillo/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type ContestHandler struct {
	db   *sqlx.DB
	tmpl *TemplateRenderer
}

func NewContestHandler(db *sqlx.DB, tmpl *TemplateRenderer) *ContestHandler {
	return &ContestHandler{db: db, tmpl: tmpl}
}

func (h *ContestHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	_, err := h.db.Exec(`INSERT INTO contests (name, sent_exch, rcvd_exch, freq, mode,
		category_operator, category_band, category_mode, category_power, category_station, category_assisted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.FormValue("name"),
		r.FormValue("sent_exch"),
		r.FormValue("rcvd_exch"),
		r.FormValue("freq"),
		r.FormValue("mode"),
		r.FormValue("category_operator"),
		r.FormValue("category_band"),
		r.FormValue("category_mode"),
		r.FormValue("category_power"),
		r.FormValue("category_station"),
		r.FormValue("category_assisted"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTable(w)
}

func (h *ContestHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	r.ParseForm()

	_, err := h.db.Exec(`UPDATE contests SET name=?, sent_exch=?, rcvd_exch=?, freq=?, mode=?,
		category_operator=?, category_band=?, category_mode=?, category_power=?, category_station=?, category_assisted=?
		WHERE id=?`,
		r.FormValue("name"),
		r.FormValue("sent_exch"),
		r.FormValue("rcvd_exch"),
		r.FormValue("freq"),
		r.FormValue("mode"),
		r.FormValue("category_operator"),
		r.FormValue("category_band"),
		r.FormValue("category_mode"),
		r.FormValue("category_power"),
		r.FormValue("category_station"),
		r.FormValue("category_assisted"),
		id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTable(w)
}

func (h *ContestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM contests WHERE id = ?", id)
	h.renderTable(w)
}

func (h *ContestHandler) GetJSON(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idNum, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var contest models.Contest
	if err := h.db.Get(&contest, "SELECT * FROM contests WHERE id = ?", idNum); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contest)
}

func (h *ContestHandler) renderTable(w http.ResponseWriter) {
	var contests []models.Contest
	h.db.Select(&contests, "SELECT * FROM contests ORDER BY name")

	w.Header().Set("Content-Type", "text/html")
	h.tmpl.RenderPartial(w, "partials/contest_table.html", map[string]any{
		"Contests": contests,
		"Modes":    modeList,
		"Freqs":    freqList,
	})
}
