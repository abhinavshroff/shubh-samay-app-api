package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetConfig — public endpoint, called by every app on launch.
// Returns flat { flags: { key: bool } } map for fast client-side merge.
func GetConfig(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			"SELECT key, enabled FROM feature_flags")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusOK, gin.H{"flags": flags})
	}
}

// ListFlags — admin only, returns full flag metadata
func ListFlags(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			"SELECT key, enabled, tier, updated_at FROM feature_flags ORDER BY key")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var out []gin.H
		for rows.Next() {
			var key, tier string
			var enabled bool
			var updated string
			rows.Scan(&key, &enabled, &tier, &updated)
			out = append(out, gin.H{"key": key, "enabled": enabled, "tier": tier, "updatedAt": updated})
		}
		c.JSON(http.StatusOK, gin.H{"flags": out})
	}
}

// UpdateFlag — admin only, flips a flag remotely.
// Every connected app picks up the change on next /config call (typically next launch
// or background refresh).
func UpdateFlag(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		_, err := pool.Exec(context.Background(),
			"UPDATE feature_flags SET enabled=$1, updated_at=NOW() WHERE key=$2",
			body.Enabled, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"key": key, "enabled": body.Enabled})
	}
}
