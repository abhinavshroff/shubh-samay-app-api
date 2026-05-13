package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/shubh-samay/api/internal/panchang"
)

func GetPanchang(w http.ResponseWriter, r *http.Request) {
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid lat")
		return
	}
	lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid lon")
		return
	}
	tz := r.URL.Query().Get("tz")
	if tz == "" {
		tz = "Asia/Kolkata"
	}

	dateStr := r.URL.Query().Get("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
	}

	result, err := panchang.Compute(date, lat, lon, tz)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// Festivals — for MVP, return a curated stub. Replace with DB query when seeded.
func GetFestivals(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, JSONMap{
		"items": []JSONMap{
			{"date": "12 May", "name": "Buddha Purnima", "nameHi": "बुद्ध पूर्णिमा", "tithiHi": "वैशाख पूर्णिमा", "daysAway": 10},
			{"date": "27 May", "name": "Amavasya", "nameHi": "अमावस्या", "tithiHi": "पितृ तर्पण का दिन", "daysAway": 25},
			{"date": "11 Jun", "name": "Nirjala Ekadashi", "nameHi": "निर्जला एकादशी", "tithiHi": "ज्येष्ठ शुक्ल एकादशी", "daysAway": 40},
		},
	})
}

func GetLunarDays(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, JSONMap{
		"items": []JSONMap{
			{"date": "8 May", "type": "ekadashi", "label": "Ekadashi"},
			{"date": "12 May", "type": "purnima", "label": "Purnima"},
			{"date": "23 May", "type": "ekadashi", "label": "Ekadashi"},
			{"date": "27 May", "type": "amavasya", "label": "Amavasya"},
			{"date": "28 May", "type": "pradosh", "label": "Pradosh"},
		},
	})
}

func FindMuhurat(w http.ResponseWriter, r *http.Request) {
	activity := r.URL.Query().Get("activity")
	if activity == "" {
		activity = "travel"
	}
	WriteJSON(w, http.StatusOK, JSONMap{
		"activity": activity,
		"date":     "Mon, 4 May",
		"time":     "7:30 AM – 9:00 AM",
		"meta":     "Rohini nakshatra • Shubha yoga • After sunrise",
	})
}
