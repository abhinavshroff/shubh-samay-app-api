package handlers

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseRequiredFloatQueryAcceptsCoordinateAliases(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/panchang?latitude=28.6139&lng=77.2090", nil)

	lat, err := parseRequiredFloatQuery(req, "lat", "latitude")
	if err != nil {
		t.Fatalf("expected latitude alias to parse: %v", err)
	}
	if lat != 28.6139 {
		t.Fatalf("unexpected latitude: got %v", lat)
	}

	lon, err := parseRequiredFloatQuery(req, "lon", "lng", "longitude")
	if err != nil {
		t.Fatalf("expected lng alias to parse: %v", err)
	}
	if lon != 77.2090 {
		t.Fatalf("unexpected longitude: got %v", lon)
	}
}

func TestFirstQueryValueAcceptsRegionalCalendarAliases(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/lunar-days?regionalCalendar=south", nil)

	got := firstQueryValue(req, "calendar", "regionalCalendar", "region", "calendarRegion")
	if got != "south" {
		t.Fatalf("unexpected regional calendar: got %q", got)
	}
}

func TestDedupeSeededFestivalsKeepsFirstDateNamePair(t *testing.T) {
	items := []seededFestival{
		{ISODate: "2026-09-04", Name: "Krishna Janmashtami", Date: "4 Sep", DaysAway: 1},
		{ISODate: "2026-09-04", Name: "Krishna Janmashtami", Date: "4 Sep", DaysAway: 1, Region: "duplicate"},
		{ISODate: "2026-09-14", Name: "Ganesh Chaturthi", Date: "14 Sep", DaysAway: 11},
	}

	got := dedupeSeededFestivals(items)

	if len(got) != 2 {
		t.Fatalf("expected 2 deduped festivals, got %d: %#v", len(got), got)
	}
	if got[0].Region != "" {
		t.Fatalf("expected first duplicate to be preserved, got %#v", got[0])
	}
	if got[0].Name != "Krishna Janmashtami" || got[1].Name != "Ganesh Chaturthi" {
		t.Fatalf("dedupe changed festival order: %#v", got)
	}
}

func TestDateWindowFromQueryUsesRequestedCalendarYear(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/lunar-days?year=2026", nil)

	from, days, rangeMode, err := dateWindowFromQuery(req, "Asia/Kolkata", 45)
	if err != nil {
		t.Fatalf("dateWindowFromQuery returned error: %v", err)
	}

	if got := from.Format("2006-01-02"); got != "2026-01-01" {
		t.Fatalf("from = %q, want 2026-01-01", got)
	}
	if days != 365 {
		t.Fatalf("days = %d, want 365", days)
	}
	if rangeMode != calendarYearRange {
		t.Fatalf("range = %q, want %q", rangeMode, calendarYearRange)
	}
}

func TestDateWindowFromQueryUsesLeapYearLength(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/festivals?calendarYear=2028", nil)

	from, days, rangeMode, err := dateWindowFromQuery(req, "Asia/Kolkata", 120)
	if err != nil {
		t.Fatalf("dateWindowFromQuery returned error: %v", err)
	}

	if got := from.Format("2006-01-02"); got != "2028-01-01" {
		t.Fatalf("from = %q, want 2028-01-01", got)
	}
	if days != 366 {
		t.Fatalf("days = %d, want 366", days)
	}
	if rangeMode != calendarYearRange {
		t.Fatalf("range = %q, want %q", rangeMode, calendarYearRange)
	}
}

func TestDateWindowFromQueryKeepsRollingWindowByDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/festivals?from=2026-05-16&days=10", nil)

	from, days, rangeMode, err := dateWindowFromQuery(req, "Asia/Kolkata", 120)
	if err != nil {
		t.Fatalf("dateWindowFromQuery returned error: %v", err)
	}

	if got := from.Format("2006-01-02"); got != "2026-05-16" {
		t.Fatalf("from = %q, want 2026-05-16", got)
	}
	if days != 10 {
		t.Fatalf("days = %d, want 10", days)
	}
	if rangeMode != "rolling" {
		t.Fatalf("range = %q, want rolling", rangeMode)
	}
}

func TestDateWindowFromQuerySupportsMobileCalendarFlag(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/lunar-days?mobileCalendar=true", nil)

	from, days, rangeMode, err := dateWindowFromQuery(req, "Asia/Kolkata", 45)
	if err != nil {
		t.Fatalf("dateWindowFromQuery returned error: %v", err)
	}

	if got := from.Format("01-02"); got != "01-01" {
		t.Fatalf("from month/day = %q, want 01-01", got)
	}
	if days != 365 && days != 366 {
		t.Fatalf("days = %d, want a whole calendar year", days)
	}
	if rangeMode != calendarYearRange {
		t.Fatalf("range = %q, want %q", rangeMode, calendarYearRange)
	}
}

func TestCalendarDataWindowFromQueryDefaultsToCurrentCalendarYear(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/lunar-days", nil)

	from, days, rangeMode, err := calendarDataWindowFromQuery(req, "Asia/Kolkata", 45)
	if err != nil {
		t.Fatalf("calendarDataWindowFromQuery returned error: %v", err)
	}

	currentYear := time.Now().In(from.Location()).Year()
	if got := from.Format("2006-01-02"); got != time.Date(currentYear, time.January, 1, 0, 0, 0, 0, from.Location()).Format("2006-01-02") {
		t.Fatalf("from = %q, want current year start", got)
	}
	if days != daysInCalendarYear(currentYear) {
		t.Fatalf("days = %d, want %d", days, daysInCalendarYear(currentYear))
	}
	if rangeMode != calendarYearRange {
		t.Fatalf("range = %q, want %q", rangeMode, calendarYearRange)
	}
}

func TestCalendarDataWindowFromQueryPreservesExplicitRollingWindow(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/lunar-days?from=2026-05-16&days=45", nil)

	from, days, rangeMode, err := calendarDataWindowFromQuery(req, "Asia/Kolkata", 45)
	if err != nil {
		t.Fatalf("calendarDataWindowFromQuery returned error: %v", err)
	}

	if got := from.Format("2006-01-02"); got != "2026-05-16" {
		t.Fatalf("from = %q, want 2026-05-16", got)
	}
	if days != 45 {
		t.Fatalf("days = %d, want 45", days)
	}
	if rangeMode != "rolling" {
		t.Fatalf("range = %q, want rolling", rangeMode)
	}
}

func TestSeededFestivalDaysAwayUsesCalendarDates(t *testing.T) {
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	date := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)

	got := calendarDaysBetween(from, date)
	if got != 1 {
		t.Fatalf("calendar date daysAway = %d, want 1", got)
	}
}
