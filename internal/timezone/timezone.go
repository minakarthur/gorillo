package timezone

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseOffsetMinutes parses a UTC offset string like "+05:30", "-08:00", "UTC".
// Returns the offset in minutes and whether parsing succeeded.
func ParseOffsetMinutes(tz string) (int, bool) {
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

// FormatOffset formats an offset in minutes as "+HH:MM" or "-HH:MM".
func FormatOffset(totalMinutes int) string {
	sign := "+"
	if totalMinutes < 0 {
		sign = "-"
		totalMinutes = -totalMinutes
	}
	return fmt.Sprintf("%s%02d:%02d", sign, totalMinutes/60, totalMinutes%60)
}

// Location converts a timezone string (UTC offset or IANA name) to *time.Location.
func Location(tz string) *time.Location {
	if min, ok := ParseOffsetMinutes(tz); ok {
		if min == 0 {
			return time.UTC
		}
		return time.FixedZone(FormatOffset(min), min*60)
	}
	loc, err := time.LoadLocation(strings.TrimSpace(tz))
	if err != nil {
		return time.UTC
	}
	return loc
}
