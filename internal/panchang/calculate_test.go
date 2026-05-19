package panchang

import (
	"encoding/json"
	"testing"
	"time"
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
		15:              MonthAshwin,       // Bharani remains Ashwin
		25:              MonthAshwin,       // late Bharani must not roll into Kartika
		30:              MonthKartika,      // Krittika full Moon
		56 + 40.0/60.0:  MonthMargashirsha, // Mrigashira full Moon
		133 + 25.0/60.0: MonthPhalguna,     // early Purva Phalguni must not stay in Magha
		180:             MonthChaitra,      // Chitra full Moon
		206 + 40.0/60.0: MonthVaishakha,    // Vishakha full Moon
		253 + 25.0/60.0: MonthAshadha,      // early Purva Ashadha must not stay in Jyeshtha
		320 + 5.0/60.0:  MonthBhadrapada,   // early Purva Bhadrapada must not stay in Shravana
		333 + 20.0/60.0: MonthBhadrapada,   // Bhadrapada full Moon
		359:             MonthAshwin,       // wrap around to Ashwini
	}

	for moonLon, want := range cases {
		if got := lunarMonthFromMoonLongitude(moonLon); got != want {
			t.Fatalf("lunarMonthFromMoonLongitude(%v) = %d, want %d", moonLon, got, want)
		}
	}
}

func TestNewFestivalAppliesThirtyDayDateAdjustment(t *testing.T) {
	loc := time.FixedZone("IST", 5*60*60+30*60)
	day := time.Date(2026, time.September, 20, 12, 0, 0, 0, loc)
	rule := festivalRule{name: "Dussehra", month: MonthAshwin, significance: "Ashwin Shukla Dashami."}
	info := tithiInfo{Name: "Dashami", NameHi: "दशमी", Paksha: "Shukla", PakshaHi: "शुक्ल"}

	got := newFestival(day, 8, rule, info, day, loc)

	if got.ISODate != "2026-10-20" {
		t.Fatalf("ISODate = %q, want 2026-10-20", got.ISODate)
	}
	if got.Date != "20 Oct" {
		t.Fatalf("Date = %q, want 20 Oct", got.Date)
	}
	if got.DaysAway != 38 {
		t.Fatalf("DaysAway = %d, want 38", got.DaysAway)
	}
}

func TestLocalMidnightUTCUsesLocationDayBoundary(t *testing.T) {
	kolkata := time.FixedZone("Asia/Kolkata", 5*60*60+30*60)
	day := time.Date(2026, time.May, 19, 12, 0, 0, 0, kolkata)

	got := localMidnightUTC(day)
	want := time.Date(2026, time.May, 18, 18, 30, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Fatalf("localMidnightUTC() = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestLocalMidnightUTCIsStableAcrossLongitudes(t *testing.T) {
	cases := []struct {
		name string
		loc  *time.Location
		want time.Time
	}{
		{
			name: "east_of_utc",
			loc:  time.FixedZone("UTC+8", 8*60*60),
			want: time.Date(2026, time.January, 1, 16, 0, 0, 0, time.UTC),
		},
		{
			name: "west_of_utc",
			loc:  time.FixedZone("UTC-7", -7*60*60),
			want: time.Date(2026, time.January, 2, 7, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		day := time.Date(2026, time.January, 2, 12, 0, 0, 0, tc.loc)
		if got := localMidnightUTC(day); !got.Equal(tc.want) {
			t.Fatalf("%s: localMidnightUTC() = %s, want %s", tc.name, got.Format(time.RFC3339), tc.want.Format(time.RFC3339))
		}
	}
}
