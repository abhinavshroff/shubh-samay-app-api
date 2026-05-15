package handlers

import (
	"database/sql"
	"time"

	"net/http"
	"strconv"

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

		from := time.Now()
		if fromStr := firstQueryValue(r, "from", "date"); fromStr != "" {
			from, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "invalid from date format, use YYYY-MM-DD")
				return
			}
		}

		days := 120
		if daysStr := r.URL.Query().Get("days"); daysStr != "" {
			days, err = strconv.Atoi(daysStr)
			if err != nil || days <= 0 {
				WriteError(w, http.StatusBadRequest, "invalid days")
				return
			}
		}

		calendar := firstQueryValue(r, "calendar", "regionalCalendar", "region", "calendarRegion")
		seeded, seedErr := seededFestivals(pool, from, days)
		if seedErr == nil && len(seeded) > 0 {
			WriteJSON(w, http.StatusOK, JSONMap{
				"calendar": panchang.NormalizeCalendar(calendar),
				"source":   "db_seed",
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
		item.Date = date.Format("2 Jan")
		item.ISODate = date.Format("2006-01-02")
		item.DaysAway = int(date.Sub(time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())).Hours() / 24)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dedupeSeededFestivals(items), nil
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
