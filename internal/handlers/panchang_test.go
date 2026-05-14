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
