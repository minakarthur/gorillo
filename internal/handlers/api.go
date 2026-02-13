package handlers

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
)

type APIHandler struct {
	db *sqlx.DB
}

func NewAPIHandler(db *sqlx.DB) *APIHandler {
	return &APIHandler{db: db}
}

func (h *APIHandler) CallsignSuggest(w http.ResponseWriter, r *http.Request) {
	q := strings.ToUpper(r.URL.Query().Get("q"))
	if len(q) < 2 {
		w.Write([]byte(""))
		return
	}

	var calls []string
	h.db.Select(&calls, `
		SELECT DISTINCT callsign FROM (
			SELECT sent_call AS callsign FROM qso_logs WHERE sent_call LIKE ?
			UNION
			SELECT rcvd_call AS callsign FROM qso_logs WHERE rcvd_call LIKE ?
		) t ORDER BY callsign LIMIT 10`, q+"%", q+"%")

	if len(calls) == 0 {
		w.Write([]byte(""))
		return
	}

	var sb strings.Builder
	for _, c := range calls {
		sb.WriteString(`<div class="suggest-item px-3 py-1 cursor-pointer hover:bg-blue-100" data-value="`)
		sb.WriteString(c)
		sb.WriteString(`">`)
		sb.WriteString(c)
		sb.WriteString(`</div>`)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(sb.String()))
}
