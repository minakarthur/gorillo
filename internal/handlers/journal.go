package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/engine/gorillo/internal/i18n"
	"github.com/engine/gorillo/internal/models"
	"github.com/jmoiron/sqlx"
)

type JournalHandler struct {
	db          *sqlx.DB
	tmpl        *TemplateRenderer
	recentCount int
}

func NewJournalHandler(db *sqlx.DB, tmpl *TemplateRenderer, recentCount int) *JournalHandler {
	return &JournalHandler{db: db, tmpl: tmpl, recentCount: recentCount}
}

func (h *JournalHandler) Page(w http.ResponseWriter, r *http.Request) {
	var owner models.Owner
	h.db.Get(&owner, "SELECT id, callsign, name, email, grid_locator, location, club, address, operators, timezone, language FROM owners WHERE id = 1")

	contestID, _ := strconv.Atoi(r.URL.Query().Get("contest_id"))

	var recent []models.QSO
	if contestID > 0 {
		h.db.Select(&recent, "SELECT * FROM qso_logs WHERE contest_id = ? ORDER BY date DESC, time DESC LIMIT ?", contestID, h.recentCount)
	} else {
		h.db.Select(&recent, "SELECT * FROM qso_logs ORDER BY date DESC, time DESC LIMIT ?", h.recentCount)
	}

	var contests []models.Contest
	h.db.Select(&contests, "SELECT * FROM contests ORDER BY name")

	loc := h.tmpl.Location()
	_, tzOffset := time.Now().In(loc).Zone()

	lang := i18n.ParseLang(owner.Language)

	h.tmpl.Render(w, "journal.html", map[string]any{
		"Owner":     owner,
		"Recent":    recent,
		"Active":    "journal",
		"Modes":     modeList,
		"Freqs":     freqList,
		"TZOffset":  tzOffset / 60,
		"Contests":  contests,
		"ContestID": contestID,
		"T":         i18n.TFunc(lang),
		"Lang":      string(lang),
	})
}

func (h *JournalHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	loc := h.tmpl.Location()

	dateStr := r.FormValue("date")
	timeStr := strings.ReplaceAll(r.FormValue("time"), ":", "")
	if len(timeStr) > 4 {
		timeStr = timeStr[:4]
	}

	var utcDate time.Time
	var utcTime string
	if dateStr != "" && len(timeStr) == 4 {
		localDT, err := time.ParseInLocation("2006-01-02 1504", dateStr+" "+timeStr, loc)
		if err == nil {
			utcDT := localDT.UTC()
			utcDate = utcDT
			utcTime = utcDT.Format("1504")
		}
	} else if dateStr != "" {
		utcDate, _ = time.Parse("2006-01-02", dateStr)
		utcTime = timeStr
	}

	var contestID *int
	if cidStr := r.FormValue("contest_id"); cidStr != "" {
		if cid, err := strconv.Atoi(cidStr); err == nil && cid > 0 {
			contestID = &cid
		}
	}

	qso := models.QSO{
		Freq:          r.FormValue("freq"),
		Mode:          r.FormValue("mode"),
		Date:          utcDate,
		Time:          utcTime,
		SentCall:      strings.ToUpper(r.FormValue("sent_call")),
		SentRST:       r.FormValue("sent_rst"),
		SentExch:      r.FormValue("sent_exch"),
		RcvdCall:      strings.ToUpper(r.FormValue("rcvd_call")),
		RcvdRST:       r.FormValue("rcvd_rst"),
		RcvdExch:      r.FormValue("rcvd_exch"),
		TransmitterID: 0,
		ContestID:     contestID,
	}

	if r.FormValue("transmitter_id") == "1" {
		qso.TransmitterID = 1
	}

	errors := map[string]bool{}
	if qso.Freq == "" {
		errors["freq"] = true
	}
	if qso.Mode == "" {
		errors["mode"] = true
	}
	if dateStr == "" || utcDate.IsZero() {
		errors["date"] = true
	}
	if utcTime == "" {
		errors["time"] = true
	}
	if qso.SentCall == "" {
		errors["sent_call"] = true
	}
	if qso.RcvdCall == "" {
		errors["rcvd_call"] = true
	}

	if len(errors) > 0 {
		var owner models.Owner
		h.db.Get(&owner, "SELECT id, callsign, name, email, grid_locator, location, club, address, operators, timezone, language FROM owners WHERE id = 1")
		var recent []models.QSO
		h.db.Select(&recent, "SELECT * FROM qso_logs ORDER BY date DESC, time DESC LIMIT ?", h.recentCount)

		var contests []models.Contest
		h.db.Select(&contests, "SELECT * FROM contests ORDER BY name")

		_, off := time.Now().In(loc).Zone()
		lang := i18n.ParseLang(owner.Language)
		cid := 0
		if contestID != nil {
			cid = *contestID
		}
		h.tmpl.Render(w, "journal.html", map[string]any{
			"Owner":     owner,
			"Recent":    recent,
			"Active":    "journal",
			"Modes":     modeList,
			"Freqs":     freqList,
			"QSO":       qso,
			"Errors":    errors,
			"TZOffset":  off / 60,
			"Contests":  contests,
			"ContestID": cid,
			"T":         i18n.TFunc(lang),
			"Lang":      string(lang),
		})
		return
	}

	result, err := h.db.Exec(`INSERT INTO qso_logs (freq, mode, date, time, sent_call, sent_rst, sent_exch, rcvd_call, rcvd_rst, rcvd_exch, transmitter_id, contest_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		qso.Freq, qso.Mode, qso.Date, qso.Time,
		qso.SentCall, qso.SentRST, qso.SentExch,
		qso.RcvdCall, qso.RcvdRST, qso.RcvdExch,
		qso.TransmitterID, qso.ContestID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		id, _ := result.LastInsertId()
		qso.ID = id
		var owner models.Owner
		h.db.Get(&owner, "SELECT callsign FROM owners WHERE id = 1")
		w.Header().Set("Content-Type", "text/html")
		h.tmpl.RenderPartial(w, "partials/qso_row.html", map[string]any{
			"QSO":       qso,
			"OwnerCall": owner.Callsign,
		})
		return
	}
	http.Redirect(w, r, "/journal", http.StatusSeeOther)
}

var modeList = []string{"CW", "PH", "FM", "RY", "DG"}

var freqList = []string{
	"1800", "3500", "7000", "14000", "21000", "28000",
	"50", "70", "144", "222", "432", "902",
	"1.2G", "2.3G", "3.4G", "5.7G", "10G",
}
