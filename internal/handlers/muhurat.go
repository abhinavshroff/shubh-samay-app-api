package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/shubh-samay/api/internal/panchang"
)

func GetMuhurat(w http.ResponseWriter, r *http.Request) {
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
	if fromStr := firstQueryValue(r, "from", "date"); fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid from/date format, use YYYY-MM-DD")
			return
		}
	}

	days := 0
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid days")
			return
		}
	}

	limit := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
	}

	activity := panchang.NormalizeMuhuratActivity(r.URL.Query().Get("activity"))
	items, err := panchang.FindMuhurat(panchang.MuhuratRequest{
		Activity: activity,
		From:     from,
		Days:     days,
		Limit:    limit,
		Lat:      lat,
		Lon:      lon,
		Timezone: tz,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, JSONMap{
		"activity":            activity,
		"supportedActivities": panchang.SupportedMuhuratActivities(),
		"items":               items,
	})
}
