package panchang

import (
	"encoding/json"
	"testing"
)

func TestResultJSONUsesFrontendKeys(t *testing.T) {
	result := Result{
		Date:      "2026-05-14",
		Weekday:   "Thursday",
		WeekdayHi: "गुरुवार",
		Tithi:     Element{Name: "Trayodashi", NameHi: "त्रयोदशी"},
		Nakshatra: Element{Name: "Ashwini"},
		Yoga:      Element{Name: "Shubha"},
		Karana:    Element{Name: "Bava"},
		Sunrise:   "5:30 AM",
		Sunset:    "6:30 PM",
		RahuKaal:  TimeRange{From: "1:30 PM", To: "3:00 PM"},
	}

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	for _, key := range []string{"date", "weekday", "weekdayHi", "tithi", "nakshatra", "yoga", "karana", "sunrise", "sunset", "rahuKaal", "gulikaKaal", "yamagandam", "abhijit", "brahmaMuhurat"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("expected JSON key %q in %s", key, payload)
		}
	}
	for _, oldKey := range []string{"Date", "Weekday", "RahuKaal", "BrahmaMuhurat"} {
		if _, ok := decoded[oldKey]; ok {
			t.Fatalf("did not expect old exported JSON key %q in %s", oldKey, payload)
		}
	}
}

func TestNormalizeCalendarSupportsRegionalAliases(t *testing.T) {
	cases := map[string]string{
		"":             CalendarNorth,
		"North":        CalendarNorth,
		"purnimanta":   CalendarNorth,
		"South":        CalendarSouth,
		"south_indian": CalendarSouth,
		"amanta":       CalendarSouth,
		"unknown":      CalendarNorth,
	}

	for input, want := range cases {
		if got := NormalizeCalendar(input); got != want {
			t.Fatalf("NormalizeCalendar(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestLunarDayJSONIncludesRegionalCalendarFields(t *testing.T) {
	item := LunarDay{
		Date:     "2026-05-27",
		Type:     "amavasya",
		Label:    "Amavasya",
		Tithi:    "Krishna Amavasya",
		TithiHi:  "कृष्ण अमावस्या",
		Paksha:   "Krishna",
		PakshaHi: "कृष्ण",
		Calendar: CalendarSouth,
		Rule:     "active at local sunrise (South Indian amanta calendar)",
		DaysAway: 13,
	}

	payload, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	for _, key := range []string{"date", "type", "label", "tithi", "tithiHi", "paksha", "pakshaHi", "calendar", "rule", "daysAway"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("expected JSON key %q in %s", key, payload)
		}
	}
	if decoded["calendar"] != CalendarSouth {
		t.Fatalf("expected calendar %q in %s", CalendarSouth, payload)
	}
}

func TestLunarMonthFromMoonLongitudeUsesFullMoonNakshatra(t *testing.T) {
	cases := map[float64]int{
		0:               MonthAshwin,       // Ashwini full Moon
		15:              MonthAshwin,       // Bharani remains closer to Ashwini than Krittika
		30:              MonthKartika,      // Krittika full Moon
		56 + 40.0/60.0:  MonthMargashirsha, // Mrigashira full Moon
		180:             MonthChaitra,      // Chitra full Moon
		206 + 40.0/60.0: MonthVaishakha,    // Vishakha full Moon
		333 + 20.0/60.0: MonthBhadrapada,   // Bhadrapada full Moon
		359:             MonthAshwin,       // wrap around to Ashwini
	}

	for moonLon, want := range cases {
		if got := lunarMonthFromMoonLongitude(moonLon); got != want {
			t.Fatalf("lunarMonthFromMoonLongitude(%v) = %d, want %d", moonLon, got, want)
		}
	}
}
