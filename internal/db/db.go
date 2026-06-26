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
  significance TEXT,
  slug        TEXT,
  year        INTEGER,
  source      TEXT
);
CREATE INDEX IF NOT EXISTS idx_festivals_date ON festivals(date);
ALTER TABLE festivals ADD COLUMN IF NOT EXISTS slug TEXT;
ALTER TABLE festivals ADD COLUMN IF NOT EXISTS year INTEGER;
ALTER TABLE festivals ADD COLUMN IF NOT EXISTS source TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_festivals_year_slug_region
  ON festivals(year, slug, region)
  WHERE slug IS NOT NULL AND year IS NOT NULL;

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

-- Seed curated Hindu festival dates for 2026 from BankBazaar's Hindu
-- festivals list. The festivals endpoint serves these database rows first so
-- admin-corrected dates are authoritative, and only falls back to computed
-- rules when no rows exist for a date window.
DELETE FROM festivals
WHERE date >= '2026-01-01'::DATE
  AND date < '2027-01-01'::DATE
  AND (slug IS NULL OR year IS NULL);

INSERT INTO festivals (date, name_en, name_hi, tithi_hi, region, significance, slug, year, source)
SELECT date::DATE, name_en, name_hi, tithi_hi, region, significance, slug, 2026, source
FROM (VALUES
  ('2026-01-14', 'Makar Sankranti', 'मकर संक्रांति', '', 'national', 'Solar transition into Makara; harvest festival.', 'makar-sankranti', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-01-14', 'Pongal', 'पोंगल', '', 'tamil-nadu', 'Tamil harvest festival.', 'pongal', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-01-14', 'Magh Bihu', 'माघ बिहू', '', 'assam', 'Assamese harvest festival.', 'magh-bihu', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-01-23', 'Vasant Panchami', 'वसंत पंचमी', 'माघ शुक्ल पंचमी', 'national', 'Magha Shukla Panchami.', 'vasant-panchami', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-02-01', 'Thaipusam', 'थाइपुसम', '', 'tamil-nadu', 'Festival dedicated to Lord Murugan.', 'thaipusam', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-02-15', 'Maha Shivaratri', 'महाशिवरात्रि', 'फाल्गुन कृष्ण चतुर्दशी', 'national', 'Phalguna Krishna Chaturdashi.', 'maha-shivaratri', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-03-03', 'Holika Dahan', 'होलिका दहन', '', 'national', 'Eve of Holi.', 'holika-dahan', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-03-04', 'Holi', 'होली', 'फाल्गुन पूर्णिमा', 'national', 'Festival of colours.', 'holi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-03-19', 'Chaitra Sukhladi', 'चैत्र शुक्लादि', '', 'national', 'Start of Chaitra month observance.', 'chaitra-sukhladi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-03-19', 'Cheti Chand', 'चेटी चंड', '', 'sindhi', 'Sindhi New Year.', 'cheti-chand', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-03-19', 'Ugadi', 'उगादी', '', 'andhra-telangana-karnataka', 'Deccan lunar new year.', 'ugadi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-03-19', 'Gudi Padwa', 'गुड़ी पड़वा', '', 'maharashtra', 'Marathi New Year.', 'gudi-padwa', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-03-26', 'Rama Navami', 'राम नवमी', 'चैत्र शुक्ल नवमी', 'national', 'Chaitra Shukla Navami.', 'rama-navami', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-02', 'Hanuman Jayanti', 'हनुमान जयंती', '', 'national', 'Birth anniversary of Lord Hanuman.', 'hanuman-jayanti', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-10', 'Hindi New Year', 'हिंदू नव वर्ष', '', 'national', 'Hindi calendar new year observance.', 'hindi-new-year', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-14', 'Vaisakhi', 'वैसाखी', '', 'punjab', 'Harvest and new year festival.', 'vaisakhi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-14', 'Vishu', 'विशु', '', 'kerala', 'Malayali New Year festival.', 'vishu', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-14', 'Tamil New Year', 'तमिल नव वर्ष', '', 'tamil-nadu', 'Tamil Puthandu.', 'tamil-new-year', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-15', 'Bengali New Year', 'बंगाली नव वर्ष', '', 'west-bengal', 'Pohela Boishakh.', 'bengali-new-year', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-15', 'Bohag Bihu', 'बोहाग बिहू', '', 'assam', 'Assamese New Year.', 'bohag-bihu', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-15', 'Vaisakhadi', 'वैशाखड़ी', '', 'regional', 'Regional Vaisakha observance.', 'vaisakhadi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-04-19', 'Akshaya Tritiya', 'अक्षय तृतीया', 'वैशाख शुक्ल तृतीया', 'national', 'Vaishakha Shukla Tritiya.', 'akshaya-tritiya', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-05-16', 'Savitri Pooja', 'सावित्री पूजा', '', 'regional', 'Savitri Puja observance.', 'savitri-pooja', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-07-16', 'Jagannath Rath Yatra', 'जगन्नाथ रथ यात्रा', '', 'odisha', 'Jagannath Rath Yatra.', 'jagannath-rath-yatra', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-07-25', 'Ashadhi Ekadashi', 'आषाढ़ी एकादशी', '', 'maharashtra', 'Ashadha Shukla Ekadashi.', 'ashadhi-ekadashi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-07-29', 'Guru Purnima', 'गुरु पूर्णिमा', 'आषाढ़ पूर्णिमा', 'national', 'Ashadha Purnima.', 'guru-purnima', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-08-15', 'Hariyali Teej', 'हरियाली तीज', '', 'north', 'Shravana Teej observance.', 'hariyali-teej', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-08-17', 'Nag Panchami', 'नाग पंचमी', '', 'national', 'Shravana Shukla Panchami.', 'nag-panchami', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-08-26', 'Onam', 'ओणम', '', 'kerala', 'Thiruvonam festival.', 'onam', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-08-28', 'Raksha Bandhan', 'रक्षा बंधन', 'श्रावण पूर्णिमा', 'national', 'Shravana Purnima.', 'raksha-bandhan', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-08-28', 'Varalakshmi Vrat', 'वरलक्ष्मी व्रत', '', 'south', 'Vrat dedicated to Goddess Lakshmi.', 'varalakshmi-vrat', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-08-31', 'Kajari Teej', 'कजरी तीज', '', 'north', 'Bhadrapada Krishna Teej observance.', 'kajari-teej', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-09-04', 'Krishna Janmashtami', 'कृष्ण जन्माष्टमी', 'भाद्रपद कृष्ण अष्टमी', 'national', 'Bhadrapada Krishna Ashtami.', 'krishna-janmashtami', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-09-14', 'Ganesh Chaturthi', 'गणेश चतुर्थी', 'भाद्रपद शुक्ल चतुर्थी', 'national', 'Bhadrapada Shukla Chaturthi.', 'ganesh-chaturthi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-09-17', 'Vishwakarma Puja', 'विश्वकर्मा पूजा', '', 'national', 'Worship of Lord Vishwakarma.', 'vishwakarma-puja', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-09-25', 'Anant Chaturdashi', 'अनंत चतुर्दशी', '', 'national', 'Bhadrapada Shukla Chaturdashi.', 'anant-chaturdashi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-10', 'Mahalaya Amavasya', 'महालया अमावस्या', '', 'east', 'Mahalaya Amavasya.', 'mahalaya-amavasya', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-11', 'Sharad Navaratri Begins', 'शारदीय नवरात्रि प्रारंभ', 'आश्विन शुक्ल प्रतिपदा', 'national', 'Ashwin Shukla Pratipada.', 'sharad-navaratri-begins', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-17', 'Maha Saptami', 'महा सप्तमी', '', 'east', 'Durga Puja Maha Saptami.', 'maha-saptami', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-18', 'Durga Puja Ashtami', 'दुर्गा पूजा अष्टमी', '', 'east', 'Durga Ashtami.', 'durga-puja-ashtami', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-19', 'Durga Maha Navami Puja', 'दुर्गा महा नवमी पूजा', '', 'east', 'Durga Maha Navami.', 'durga-maha-navami-puja', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-20', 'Dussehra', 'दशहरा', 'आश्विन शुक्ल दशमी', 'national', 'Ashwin Shukla Dashami.', 'dussehra', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-20', 'Sharad Navaratri Parana', 'शारदीय नवरात्रि पारण', '', 'national', 'Sharad Navaratri Parana.', 'sharad-navaratri-parana', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-25', 'Sharad Purnima', 'शरद पूर्णिमा', '', 'national', 'Ashwin Purnima.', 'sharad-purnima', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-29', 'Karwa Chauth', 'करवा चौथ', 'कार्तिक कृष्ण चतुर्थी', 'north', 'Kartika Krishna Chaturthi.', 'karwa-chauth', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-10-31', 'Maharishi Valmiki Jayanti', 'महर्षि वाल्मीकि जयंती', '', 'national', 'Birth anniversary of Maharishi Valmiki.', 'maharishi-valmiki-jayanti', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-11-06', 'Dhanteras', 'धनतेरस', 'कार्तिक कृष्ण त्रयोदशी', 'national', 'Kartika Krishna Trayodashi.', 'dhanteras', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-11-08', 'Diwali', 'दीपावली', 'कार्तिक अमावस्या', 'national', 'Kartika Amavasya.', 'diwali', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-11-08', 'Naraka Chaturdashi', 'नरक चतुर्दशी', '', 'national', 'Naraka Chaturdashi.', 'naraka-chaturdashi', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-11-09', 'Govardhan Puja', 'गोवर्धन पूजा', 'कार्तिक शुक्ल प्रतिपदा', 'national', 'Kartika Shukla Pratipada.', 'govardhan-puja', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-11-11', 'Bhai Dooj', 'भाई दूज', 'कार्तिक शुक्ल द्वितीया', 'national', 'Kartika Shukla Dwitiya.', 'bhai-dooj', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-11-15', 'Chhath Puja', 'छठ पूजा', '', 'bihar', 'Chhath Puja.', 'chhath-puja', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-11-24', 'Kartik Purnima', 'कार्तिक पूर्णिमा', 'कार्तिक पूर्णिमा', 'national', 'Kartika Purnima.', 'kartik-purnima', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-12-16', 'Dhanu Sankranti', 'धनु संक्रांति', '', 'regional', 'Solar transition into Dhanu.', 'dhanu-sankranti', 'BankBazaar Hindu Festivals List 2026'),
  ('2026-12-20', 'Gita Jayanti', 'गीता जयंती', '', 'national', 'Margashirsha Shukla Ekadashi.', 'gita-jayanti', 'BankBazaar Hindu Festivals List 2026')
) AS seed(date, name_en, name_hi, tithi_hi, region, significance, slug, source)
ON CONFLICT (year, slug, region) WHERE slug IS NOT NULL AND year IS NOT NULL
DO UPDATE SET
  date = EXCLUDED.date,
  name_en = EXCLUDED.name_en,
  name_hi = EXCLUDED.name_hi,
  tithi_hi = EXCLUDED.tithi_hi,
  significance = EXCLUDED.significance,
  source = EXCLUDED.source;

`
