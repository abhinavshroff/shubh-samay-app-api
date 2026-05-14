package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/shubh-samay/api/internal/panchang"
)

func GetPanchang(w http.ResponseWriter, r *http.Request) {
	lat, err := parseRequiredFloatQuery(r, "lat", "latitude")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid lat")
		return
	}
	lon, err := parseRequiredFloatQuery(r, "lon", "lng", "longitude")
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

func GetLunarDays(w http.ResponseWriter, r *http.Request) {
	lat, err := parseRequiredFloatQuery(r, "lat", "latitude")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid lat")
		return
	}
	lon, err := parseRequiredFloatQuery(r, "lon", "lng", "longitude")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid lon")
		return
	}

	tz := r.URL.Query().Get("tz")
	if tz == "" {
		tz = "Asia/Kolkata"
	}

	from := time.Now()
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid from date format, use YYYY-MM-DD")
			return
		}
	}

	days := 45
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid days")
			return
		}
	}

	calendar := firstQueryValue(r, "calendar", "regionalCalendar", "region", "calendarRegion")
	items, err := panchang.LunarDays(from, days, lat, lon, tz, calendar)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, JSONMap{
		"calendar": panchang.NormalizeCalendar(calendar),
		"items":    items,
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

func parseRequiredFloatQuery(r *http.Request, names ...string) (float64, error) {
	query := r.URL.Query()
	for _, name := range names {
		if value := query.Get(name); value != "" {
			return strconv.ParseFloat(value, 64)
		}
	}
	return 0, strconv.ErrSyntax
}

func firstQueryValue(r *http.Request, names ...string) string {
	query := r.URL.Query()
	for _, name := range names {
		if value := query.Get(name); value != "" {
			return value
		}
	}
	return ""
}
