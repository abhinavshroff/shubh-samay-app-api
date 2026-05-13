package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// GetConfig — public endpoint, called by every app on launch.
// Returns flat { flags: { key: bool } } map for fast client-side merge.
func GetConfig(db *sql.DB) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), "SELECT key, enabled FROM feature_flags")
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		flags := make(map[string]bool)
		for rows.Next() {
			var key string
			var enabled bool
			if err := rows.Scan(&key, &enabled); err != nil {
				continue
			}
			flags[key] = enabled
		}
		if err := rows.Err(); err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, JSONMap{"flags": flags})
	}
}

// ListFlags — admin only, returns full flag metadata.
func ListFlags(db *sql.DB) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), "SELECT key, enabled, tier, updated_at FROM feature_flags ORDER BY key")
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		out := []JSONMap{}
		for rows.Next() {
			var key, tier string
			var enabled bool
			var updated time.Time
			if err := rows.Scan(&key, &enabled, &tier, &updated); err != nil {
				continue
			}
			out = append(out, JSONMap{"key": key, "enabled": enabled, "tier": tier, "updatedAt": updated.Format(time.RFC3339)})
		}
		if err := rows.Err(); err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, JSONMap{"flags": out})
	}
}

// UpdateFlag — admin only, flips a flag remotely.
// Every connected app picks up the change on next /config call (typically next launch
// or background refresh).
func UpdateFlag(db *sql.DB) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid body")
			return
		}
		_, err := db.ExecContext(r.Context(), "UPDATE feature_flags SET enabled=$1, updated_at=NOW() WHERE key=$2", body.Enabled, key)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, JSONMap{"key": key, "enabled": body.Enabled})
	}
}
