package cabrillo

import (
	"fmt"
	"io"
	"strings"

	"github.com/engine/gorillo/internal/models"
)

func Export(w io.Writer, owner *models.Owner, contest *models.Contest, qsos []models.QSO) error {
	lines := []string{"START-OF-LOG: 3.0"}

	addIfSet := func(tag, val string) {
		if val != "" {
			lines = append(lines, tag+": "+val)
		}
	}

	addIfSet("CALLSIGN", owner.Callsign)
	if contest != nil {
		addIfSet("CONTEST", contest.Name)
		addIfSet("CATEGORY-ASSISTED", contest.CategoryAssisted)
		addIfSet("CATEGORY-BAND", contest.CategoryBand)
		addIfSet("CATEGORY-MODE", contest.CategoryMode)
		addIfSet("CATEGORY-OPERATOR", contest.CategoryOperator)
		addIfSet("CATEGORY-POWER", contest.CategoryPower)
		addIfSet("CATEGORY-STATION", contest.CategoryStation)
	}
	addIfSet("NAME", owner.Name)
	addIfSet("EMAIL", owner.Email)
	addIfSet("GRID-LOCATOR", owner.GridLocator)
	addIfSet("LOCATION", owner.Location)
	addIfSet("CLUB", owner.Club)

	if owner.Address != "" {
		for _, line := range strings.Split(owner.Address, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				lines = append(lines, "ADDRESS: "+line)
			}
		}
	}

	if owner.Operators != "" {
		lines = append(lines, "OPERATORS: "+owner.Operators)
	}

	lines = append(lines, "CREATED-BY: Gorillo v1.0")
	lines = append(lines, "SOAPBOX:")

	for _, q := range qsos {
		qsoLine := fmt.Sprintf("QSO: %5s %2s %s %s %-13s %3s %-6s %-13s %3s %-6s %d",
			q.Freq, q.Mode, q.Date.Format("2006-01-02"), q.Time,
			q.SentCall, q.SentRST, q.SentExch,
			q.RcvdCall, q.RcvdRST, q.RcvdExch,
			q.TransmitterID,
		)
		lines = append(lines, qsoLine)
	}

	lines = append(lines, "END-OF-LOG:")

	_, err := io.WriteString(w, strings.Join(lines, "\r\n")+"\r\n")
	return err
}
