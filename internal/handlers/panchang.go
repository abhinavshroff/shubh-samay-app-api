package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shubh-samay/api/internal/panchang"
)

func GetPanchang(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat"})
		return
	}
	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lon"})
		return
	}
	tz := c.DefaultQuery("tz", "Asia/Kolkata")

	dateStr := c.Query("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
	}

	result, err := panchang.Compute(date, lat, lon, tz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Festivals — for MVP, return a curated stub. Replace with DB query when seeded.
func GetFestivals(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []gin.H{
			{"date": "12 May",  "name": "Buddha Purnima",   "nameHi": "बुद्ध पूर्णिमा",     "tithiHi": "वैशाख पूर्णिमा",      "daysAway": 10},
			{"date": "27 May",  "name": "Amavasya",         "nameHi": "अमावस्या",          "tithiHi": "पितृ तर्पण का दिन",   "daysAway": 25},
			{"date": "11 Jun",  "name": "Nirjala Ekadashi", "nameHi": "निर्जला एकादशी",   "tithiHi": "ज्येष्ठ शुक्ल एकादशी", "daysAway": 40},
		},
	})
}

func GetLunarDays(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []gin.H{
			{"date": "8 May",  "type": "ekadashi", "label": "Ekadashi"},
			{"date": "12 May", "type": "purnima",  "label": "Purnima"},
			{"date": "23 May", "type": "ekadashi", "label": "Ekadashi"},
			{"date": "27 May", "type": "amavasya", "label": "Amavasya"},
			{"date": "28 May", "type": "pradosh",  "label": "Pradosh"},
		},
	})
}

func FindMuhurat(c *gin.Context) {
	activity := c.DefaultQuery("activity", "travel")
	c.JSON(http.StatusOK, gin.H{
		"activity": activity,
		"date":     "Mon, 4 May",
		"time":     "7:30 AM – 9:00 AM",
		"meta":     "Rohini nakshatra • Shubha yoga • After sunrise",
	})
}
