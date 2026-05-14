package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func Connect(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}

const schemaSQL = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS feature_flags (
  key         TEXT PRIMARY KEY,
  enabled     BOOLEAN NOT NULL DEFAULT TRUE,
  tier        TEXT NOT NULL DEFAULT 'Freemium',
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS devices (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fcm_token       TEXT UNIQUE NOT NULL,
  language        TEXT NOT NULL DEFAULT 'en',
  city_name       TEXT,
  latitude        DOUBLE PRECISION,
  longitude       DOUBLE PRECISION,
  timezone        TEXT DEFAULT 'Asia/Kolkata',
  calendar_region TEXT DEFAULT 'north',
  notif_rahu      BOOLEAN DEFAULT TRUE,
  notif_morning   BOOLEAN DEFAULT TRUE,
  notif_festival  BOOLEAN DEFAULT TRUE,
  notif_ekadashi  BOOLEAN DEFAULT TRUE,
  notif_brahma    BOOLEAN DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_devices_fcm ON devices(fcm_token);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS calendar_region TEXT DEFAULT 'north';

CREATE TABLE IF NOT EXISTS festivals (
  id          SERIAL PRIMARY KEY,
  date        DATE NOT NULL,
  name_en     TEXT NOT NULL,
  name_hi     TEXT,
  name_te     TEXT,
  tithi_hi    TEXT,
  region      TEXT DEFAULT 'national',
  significance TEXT
);
CREATE INDEX IF NOT EXISTS idx_festivals_date ON festivals(date);

CREATE TABLE IF NOT EXISTS notification_log (
  id          SERIAL PRIMARY KEY,
  device_id   UUID REFERENCES devices(id) ON DELETE CASCADE,
  kind        TEXT NOT NULL,
  sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  success     BOOLEAN NOT NULL
);

-- Seed default flags (matches frontend defaults)
INSERT INTO feature_flags (key, enabled, tier) VALUES
  ('rahu_kaal_today',        TRUE, 'Free'),
  ('panchang_today',         TRUE, 'Free'),
  ('sunrise_sunset',         TRUE, 'Free'),
  ('festival_calendar',      TRUE, 'Free'),
  ('notif_rahu_kaal',        TRUE, 'Free'),
  ('abhijit_muhurat',        TRUE, 'Freemium'),
  ('brahma_muhurat',         TRUE, 'Freemium'),
  ('muhurat_travel',         TRUE, 'Freemium'),
  ('muhurat_business',       TRUE, 'Freemium'),
  ('muhurat_home',           TRUE, 'Freemium'),
  ('muhurat_vehicle',        TRUE, 'Freemium'),
  ('muhurat_marriage',       TRUE, 'Freemium'),
  ('kundali_muhurat',        TRUE, 'Freemium'),
  ('key_lunar_days',         TRUE, 'Freemium'),
  ('regional_calendar',      TRUE, 'Freemium'),
  ('notif_morning_panchang', TRUE, 'Freemium'),
  ('notif_festivals',        TRUE, 'Freemium'),
  ('notif_ekadashi',         TRUE, 'Freemium'),
  ('notif_brahma',           TRUE, 'Freemium'),
  ('pdf_export',             TRUE, 'Freemium'),
  ('home_widget',            TRUE, 'Freemium')
ON CONFLICT (key) DO NOTHING;
`
