package handlers

import (
	"net/http"
	"strconv"

	"github.com/engine/gorillo/internal/cabrillo"
	"github.com/engine/gorillo/internal/models"
	"github.com/jmoiron/sqlx"
)

type ExportHandler struct {
	db *sqlx.DB
}

func NewExportHandler(db *sqlx.DB) *ExportHandler {
	return &ExportHandler{db: db}
}

func (h *ExportHandler) Cabrillo(w http.ResponseWriter, r *http.Request) {
	var owner models.Owner
	h.db.Get(&owner, "SELECT id, callsign, name, email, grid_locator, location, club, address, operators, timezone, language FROM owners WHERE id = 1")

	contestID, _ := strconv.Atoi(r.URL.Query().Get("contest_id"))

	var contest *models.Contest
	if contestID > 0 {
		var c models.Contest
		if err := h.db.Get(&c, "SELECT * FROM contests WHERE id = ?", contestID); err == nil {
			contest = &c
		}
	}

	var qsos []models.QSO
	if contestID > 0 {
		h.db.Select(&qsos, "SELECT * FROM qso_logs WHERE contest_id = ? ORDER BY date ASC, time ASC", contestID)
	} else {
		h.db.Select(&qsos, "SELECT * FROM qso_logs ORDER BY date ASC, time ASC")
	}

	filename := "gorillo"
	if owner.Callsign != "" {
		filename = owner.Callsign
	}
	if contest != nil && contest.Name != "" {
		filename += "-" + contest.Name
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".log\"")

	cabrillo.Export(w, &owner, contest, qsos)
}
