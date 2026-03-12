package runstatus

import (
	"fmt"
	"strings"
	"time"
)

var shortMonthsByLocale = map[string][]string{
	localeRU: {"янв", "фев", "мар", "апр", "май", "июн", "июл", "авг", "сен", "окт", "ноя", "дек"},
	localeEN: {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func formatRecentStatusTimeLabel(value string, locale string, reference time.Time) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	parsed, err := time.Parse(time.RFC3339Nano, trimmed)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, trimmed)
		if err != nil {
			return ""
		}
	}
	parsed = parsed.UTC()
	reference = reference.UTC()
	if sameUTCDay(parsed, reference) {
		return parsed.Format("15:04")
	}

	normalizedLocale := normalizeLocale(locale, localeEN)
	months := shortMonthsByLocale[normalizedLocale]
	monthLabel := parsed.Month().String()
	if len(months) >= int(parsed.Month()) {
		monthLabel = months[parsed.Month()-1]
	}
	return fmt.Sprintf("%d %s %02d:%02d", parsed.Day(), monthLabel, parsed.Hour(), parsed.Minute())
}

func sameUTCDay(left time.Time, right time.Time) bool {
	left = left.UTC()
	right = right.UTC()
	return left.Year() == right.Year() && left.YearDay() == right.YearDay()
}
