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
