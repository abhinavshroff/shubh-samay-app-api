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

-- Seed curated common 2026 India festival dates. The festivals endpoint
-- serves these database rows first so admin-corrected dates are authoritative,
-- and only falls back to computed rules when no rows exist for a date window.
INSERT INTO festivals (date, name_en, name_hi, tithi_hi, region, significance)
SELECT date::DATE, name_en, name_hi, tithi_hi, region, significance
FROM (VALUES
  ('2026-01-23', 'Vasant Panchami', 'वसंत पंचमी', 'माघ शुक्ल पंचमी', 'national', 'Magha Shukla Panchami'),
  ('2026-02-15', 'Maha Shivaratri', 'महाशिवरात्रि', 'फाल्गुन कृष्ण चतुर्दशी', 'national', 'Phalguna Krishna Chaturdashi'),
  ('2026-03-04', 'Holi', 'होली', 'फाल्गुन पूर्णिमा', 'national', 'Festival of colours'),
  ('2026-03-26', 'Rama Navami', 'राम नवमी', 'चैत्र शुक्ल नवमी', 'national', 'Chaitra Shukla Navami'),
  ('2026-04-20', 'Akshaya Tritiya', 'अक्षय तृतीया', 'वैशाख शुक्ल तृतीया', 'national', 'Vaishakha Shukla Tritiya'),
  ('2026-05-01', 'Buddha Purnima', 'बुद्ध पूर्णिमा', 'वैशाख पूर्णिमा', 'national', 'Vaishakha Purnima'),
  ('2026-07-29', 'Guru Purnima', 'गुरु पूर्णिमा', 'आषाढ़ पूर्णिमा', 'national', 'Ashadha Purnima'),
  ('2026-08-28', 'Raksha Bandhan', 'रक्षा बंधन', 'श्रावण पूर्णिमा', 'national', 'Shravana Purnima'),
  ('2026-09-04', 'Krishna Janmashtami', 'कृष्ण जन्माष्टमी', 'भाद्रपद कृष्ण अष्टमी', 'national', 'Bhadrapada Krishna Ashtami'),
  ('2026-09-14', 'Ganesh Chaturthi', 'गणेश चतुर्थी', 'भाद्रपद शुक्ल चतुर्थी', 'national', 'Bhadrapada Shukla Chaturthi'),
  ('2026-10-11', 'Sharad Navaratri Begins', 'शारदीय नवरात्रि प्रारंभ', 'आश्विन शुक्ल प्रतिपदा', 'national', 'Ashwin Shukla Pratipada'),
  ('2026-10-20', 'Dussehra', 'दशहरा', 'आश्विन शुक्ल दशमी', 'national', 'Ashwin Shukla Dashami'),
  ('2026-10-29', 'Karwa Chauth', 'करवा चौथ', 'कार्तिक कृष्ण चतुर्थी', 'north', 'Kartika Krishna Chaturthi'),
  ('2026-11-06', 'Dhanteras', 'धनतेरस', 'कार्तिक कृष्ण त्रयोदशी', 'national', 'Kartika Krishna Trayodashi'),
  ('2026-11-08', 'Diwali', 'दीपावली', 'कार्तिक अमावस्या', 'national', 'Kartika Amavasya'),
  ('2026-11-10', 'Govardhan Puja', 'गोवर्धन पूजा', 'कार्तिक शुक्ल प्रतिपदा', 'national', 'Kartika Shukla Pratipada'),
  ('2026-11-11', 'Bhai Dooj', 'भाई दूज', 'कार्तिक शुक्ल द्वितीया', 'national', 'Kartika Shukla Dwitiya'),
  ('2026-11-24', 'Dev Diwali', 'देव दीपावली', 'कार्तिक पूर्णिमा', 'national', 'Kartika Purnima')
) AS seed(date, name_en, name_hi, tithi_hi, region, significance)
WHERE NOT EXISTS (
  SELECT 1 FROM festivals f WHERE f.date = seed.date::DATE AND f.name_en = seed.name_en
);

`
