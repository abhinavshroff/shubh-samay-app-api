package notifications

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shubh-samay/api/internal/panchang"
)

const fcmEndpoint = "https://fcm.googleapis.com/fcm/send"

type fcmMessage struct {
	To           string                 `json:"to"`
	Notification map[string]interface{} `json:"notification"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Priority     string                 `json:"priority"`
}

// DispatchRahuKaal — runs every minute via cron.
// For each device, computes when Rahu Kaal starts today and pushes
// a notification 15 minutes before.
func DispatchRahuKaal(pool *sql.DB, fcmKey string) {
	ctx := context.Background()
	rows, err := pool.QueryContext(ctx,
		`SELECT id, fcm_token, language, latitude, longitude, timezone
     FROM devices WHERE notif_rahu=TRUE`)
	if err != nil {
		return
	}
	defer rows.Close()

	now := time.Now()
	for rows.Next() {
		var id, token, lang, tz string
		var lat, lon float64
		if err := rows.Scan(&id, &token, &lang, &lat, &lon, &tz); err != nil {
			continue
		}

		result, err := panchang.Compute(now, lat, lon, tz)
		if err != nil {
			continue
		}

		// Parse Rahu Kaal start time
		loc, _ := time.LoadLocation(tz)
		rahuStart, err := time.ParseInLocation("3:04 PM", result.RahuKaal.From, loc)
		if err != nil {
			continue
		}
		// Reproject onto today's date
		rahuStart = time.Date(now.Year(), now.Month(), now.Day(),
			rahuStart.Hour(), rahuStart.Minute(), 0, 0, loc)

		// Notify if Rahu Kaal starts in 14–16 minutes (1-minute window)
		mins := time.Until(rahuStart).Minutes()
		if mins >= 14 && mins <= 16 {
			title, body := localiseRahu(lang, result.RahuKaal.From, result.RahuKaal.To)
			sendFCM(fcmKey, token, title, body)
			pool.ExecContext(ctx, "INSERT INTO notification_log (device_id, kind, success) VALUES ($1, 'rahu_kaal', TRUE)", id)
		}
	}
}

// DispatchMorning — runs at 7:00 AM IST daily
func DispatchMorning(pool *sql.DB, fcmKey string) {
	ctx := context.Background()
	rows, err := pool.QueryContext(ctx,
		`SELECT id, fcm_token, language, latitude, longitude, timezone
     FROM devices WHERE notif_morning=TRUE`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, token, lang, tz string
		var lat, lon float64
		if err := rows.Scan(&id, &token, &lang, &lat, &lon, &tz); err != nil {
			continue
		}
		result, err := panchang.Compute(time.Now(), lat, lon, tz)
		if err != nil {
			continue
		}
		title, body := localiseMorning(lang, result)
		sendFCM(fcmKey, token, title, body)
		pool.ExecContext(ctx, "INSERT INTO notification_log (device_id, kind, success) VALUES ($1, 'morning', TRUE)", id)
	}
}

func sendFCM(fcmKey, token, title, body string) {
	msg := fcmMessage{
		To: token,
		Notification: map[string]interface{}{
			"title": title, "body": body,
			"sound": "default", "channel_id": "panchang",
		},
		Priority: "high",
	}
	payload, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", fcmEndpoint, bytes.NewReader(payload))
	req.Header.Set("Authorization", "key="+fcmKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}

func localiseRahu(lang, from, to string) (string, string) {
	switch lang {
	case "hi":
		return "⚠️ राहु काल 15 मिनट में शुरू",
			fmt.Sprintf("राहु काल: %s – %s. शुभ कार्य न करें।", from, to)
	case "te":
		return "⚠️ రాహు కాలం 15 నిమిషాల్లో",
			fmt.Sprintf("రాహు కాలం: %s – %s. శుభ కార్యాలు చేయవద్దు.", from, to)
	default:
		return "⚠️ Rahu Kaal in 15 min",
			fmt.Sprintf("Rahu Kaal: %s – %s. Avoid auspicious activities.", from, to)
	}
}

func localiseMorning(lang string, r *panchang.Result) (string, string) {
	switch lang {
	case "hi":
		return "🌅 आज का पंचांग",
			fmt.Sprintf("%s तिथि • %s नक्षत्र • अभिजित: %s",
				r.Tithi.NameHi, r.Nakshatra.NameHi, r.Abhijit.From)
	case "te":
		return "🌅 నేటి పంచాంగం",
			fmt.Sprintf("%s తిథి • %s నక్షత్రం • అభిజిత్: %s",
				r.Tithi.NameTe, r.Nakshatra.Name, r.Abhijit.From)
	default:
		return "🌅 Today's Panchang",
			fmt.Sprintf("%s tithi • %s nakshatra • Abhijit: %s",
				r.Tithi.Name, r.Nakshatra.Name, r.Abhijit.From)
	}
}
