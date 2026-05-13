package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type registerDeviceReq struct {
	FCMToken      string  `json:"fcmToken"`
	Language      string  `json:"language"`
	CityName      string  `json:"cityName"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Timezone      string  `json:"timezone"`
	NotifRahu     bool    `json:"notifRahu"`
	NotifMorning  bool    `json:"notifMorning"`
	NotifFestival bool    `json:"notifFestival"`
	NotifEkadashi bool    `json:"notifEkadashi"`
	NotifBrahma   bool    `json:"notifBrahma"`
}

func RegisterDevice(db *sql.DB) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerDeviceReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.FCMToken == "" {
			WriteError(w, http.StatusBadRequest, "fcmToken is required")
			return
		}
		if req.Language == "" {
			req.Language = "en"
		}
		if req.Timezone == "" {
			req.Timezone = "Asia/Kolkata"
		}
		_, err := db.ExecContext(r.Context(),
			`INSERT INTO devices (fcm_token, language, city_name, latitude, longitude, timezone,
              notif_rahu, notif_morning, notif_festival, notif_ekadashi, notif_brahma)
       VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
       ON CONFLICT (fcm_token) DO UPDATE SET
         language=EXCLUDED.language, city_name=EXCLUDED.city_name,
         latitude=EXCLUDED.latitude, longitude=EXCLUDED.longitude, timezone=EXCLUDED.timezone,
         notif_rahu=EXCLUDED.notif_rahu, notif_morning=EXCLUDED.notif_morning,
         notif_festival=EXCLUDED.notif_festival, notif_ekadashi=EXCLUDED.notif_ekadashi,
         notif_brahma=EXCLUDED.notif_brahma, updated_at=NOW()`,
			req.FCMToken, req.Language, req.CityName, req.Latitude, req.Longitude, req.Timezone,
			req.NotifRahu, req.NotifMorning, req.NotifFestival, req.NotifEkadashi, req.NotifBrahma)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, JSONMap{"status": "ok"})
	}
}
