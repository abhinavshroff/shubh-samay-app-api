package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/shubh-samay/api/internal/panchang"
)

// GetFestivals returns a handler that serves curated festival rows from the
// database first, falling back to computed astronomical rules only when the
// database has no rows for the requested date window.
func GetFestivals(pool *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lat, err := optionalFloatQuery(r, 28.6139, "lat", "latitude")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid lat")
			return
		}
		lon, err := optionalFloatQuery(r, 77.2090, "lon", "lng", "longitude")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid lon")
			return
		}

		tz := r.URL.Query().Get("tz")
		if tz == "" {
			tz = "Asia/Kolkata"
		}

		from, days, rangeMode, err := calendarDataWindowFromQuery(r, tz, 120)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		calendar := firstQueryValue(r, "calendar", "regionalCalendar", "region", "calendarRegion")
		seeded, seedErr := seededFestivals(pool, from, days)
		if seedErr == nil && len(seeded) > 0 {
			WriteJSON(w, http.StatusOK, JSONMap{
				"calendar": panchang.NormalizeCalendar(calendar),
				"source":   "db_seed",
				"range":    rangeMode,
				"from":     from.Format("2006-01-02"),
				"days":     days,
				"items":    seeded,
			})
			return
		}

		items, computeErr := panchang.Festivals(from, days, lat, lon, tz, calendar)
		if computeErr != nil {
			if seedErr != nil {
				WriteError(w, http.StatusInternalServerError, seedErr.Error())
				return
			}
			WriteError(w, http.StatusInternalServerError, computeErr.Error())
			return
		}
		WriteJSON(w, http.StatusOK, JSONMap{
			"calendar": panchang.NormalizeCalendar(calendar),
			"source":   "computed",
			"range":    rangeMode,
			"from":     from.Format("2006-01-02"),
			"days":     days,
			"items":    items,
		})
	}
}

type seededFestival struct {
	Date         string `json:"date"`
	ISODate      string `json:"isoDate"`
	Name         string `json:"name"`
	NameHi       string `json:"nameHi,omitempty"`
	NameTe       string `json:"nameTe,omitempty"`
	TithiHi      string `json:"tithiHi,omitempty"`
	Region       string `json:"region,omitempty"`
	Significance string `json:"significance,omitempty"`
	DaysAway     int    `json:"daysAway"`
}

func seededFestivals(pool *sql.DB, from time.Time, days int) ([]seededFestival, error) {
	if pool == nil {
		return nil, sql.ErrConnDone
	}
	to := from.AddDate(0, 0, days)
	rows, err := pool.Query(`
SELECT DISTINCT ON (date, lower(trim(name_en)))
       date, name_en, COALESCE(name_hi, ''), COALESCE(name_te, ''), COALESCE(tithi_hi, ''), COALESCE(region, ''), COALESCE(significance, '')
FROM festivals
WHERE date >= $1 AND date < $2
ORDER BY date, lower(trim(name_en)), id`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []seededFestival{}
	for rows.Next() {
		var date time.Time
		var item seededFestival
		if err := rows.Scan(&date, &item.Name, &item.NameHi, &item.NameTe, &item.TithiHi, &item.Region, &item.Significance); err != nil {
			return nil, err
		}
		localDate := localCalendarDate(date, from.Location())
		item.Date = localDate.Format("2 Jan")
		item.ISODate = localDate.Format("2006-01-02")
		item.DaysAway = calendarDaysBetween(from, localDate)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dedupeSeededFestivals(items), nil
}

func localCalendarDate(date time.Time, loc *time.Location) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
}

func calendarDaysBetween(from, date time.Time) int {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	dateInFromLocation := localCalendarDate(date, from.Location())
	dateOnly := time.Date(dateInFromLocation.Year(), dateInFromLocation.Month(), dateInFromLocation.Day(), 0, 0, 0, 0, time.UTC)
	return int(dateOnly.Sub(fromDate).Hours() / 24)
}

func dedupeSeededFestivals(items []seededFestival) []seededFestival {
	if len(items) < 2 {
		return items
	}

	deduped := items[:0]
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		key := item.ISODate + "\x00" + item.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, item)
	}
	return deduped
}

func optionalFloatQuery(r *http.Request, fallback float64, names ...string) (float64, error) {
	query := r.URL.Query()
	for _, name := range names {
		if value := query.Get(name); value != "" {
			return strconv.ParseFloat(value, 64)
		}
	}
	return fallback, nil
}
