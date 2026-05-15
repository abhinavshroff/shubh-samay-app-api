package handlers

import (
	"net/http/httptest"
	"testing"
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
