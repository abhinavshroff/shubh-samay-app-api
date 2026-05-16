package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const calendarYearRange = "calendar_year"

// dateWindowFromQuery returns either the requested rolling date window or the
// full local calendar year when the mobile calendar asks for year-scoped data.
func dateWindowFromQuery(r *http.Request, tz string, defaultDays int) (time.Time, int, string, error) {
	return dateWindowFromQueryOptions(r, tz, defaultDays, false)
}

// calendarDataWindowFromQuery defaults an unscoped calendar feed request to the
// complete current local calendar year while preserving explicit rolling windows.
func calendarDataWindowFromQuery(r *http.Request, tz string, defaultDays int) (time.Time, int, string, error) {
	return dateWindowFromQueryOptions(r, tz, defaultDays, true)
}

func dateWindowFromQueryOptions(r *http.Request, tz string, defaultDays int, defaultToCurrentYear bool) (time.Time, int, string, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, 0, "", fmt.Errorf("invalid timezone")
	}

	if year, ok, err := calendarYearFromQuery(r, loc); ok || err != nil {
		if err != nil {
			return time.Time{}, 0, "", err
		}
		return calendarYearWindow(year, loc), daysInCalendarYear(year), calendarYearRange, nil
	}

	if defaultToCurrentYear && !hasRollingWindowQuery(r) {
		year := time.Now().In(loc).Year()
		return calendarYearWindow(year, loc), daysInCalendarYear(year), calendarYearRange, nil
	}

	from := time.Now().In(loc)
	if fromStr := firstQueryValue(r, "from", "date"); fromStr != "" {
		from, err = time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return time.Time{}, 0, "", fmt.Errorf("invalid from date format, use YYYY-MM-DD")
		}
	}

	days := defaultDays
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return time.Time{}, 0, "", fmt.Errorf("invalid days")
		}
	}

	return from, days, "rolling", nil
}

func calendarYearWindow(year int, loc *time.Location) time.Time {
	return time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
}

func hasRollingWindowQuery(r *http.Request) bool {
	if firstQueryValue(r, "from", "date", "days") != "" {
		return true
	}

	return firstQueryValue(r, "range", "scope") != ""
}

func daysInCalendarYear(year int) int {
	if time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC).YearDay() == 366 {
		return 366
	}
	return 365
}

func calendarYearFromQuery(r *http.Request, loc *time.Location) (int, bool, error) {
	if yearStr := firstQueryValue(r, "year", "calendarYear"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1 || year > 9999 {
			return 0, true, fmt.Errorf("invalid year")
		}
		return year, true, nil
	}

	scope := strings.ToLower(strings.TrimSpace(firstQueryValue(r, "range", "scope")))
	if scope == calendarYearRange || scope == "year" || scope == "current_year" || scope == "calendar-year" {
		return time.Now().In(loc).Year(), true, nil
	}

	if truthyQuery(r, "mobileCalendar", "calendarView", "fullYear") {
		return time.Now().In(loc).Year(), true, nil
	}

	return 0, false, nil
}

func truthyQuery(r *http.Request, names ...string) bool {
	query := r.URL.Query()
	for _, name := range names {
		value := strings.ToLower(strings.TrimSpace(query.Get(name)))
		if value == "1" || value == "true" || value == "yes" {
			return true
		}
	}
	return false
}
