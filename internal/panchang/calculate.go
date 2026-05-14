package panchang

import (
	"fmt"
	"math"
	"path/filepath"
	"time"

	swe "github.com/mshafiee/swephgo"
)

// Tithi names (1-15 Shukla, 16-30 Krishna)
var TithiNames = []string{
	"Pratipada", "Dwitiya", "Tritiya", "Chaturthi", "Panchami",
	"Shashthi", "Saptami", "Ashtami", "Navami", "Dashami",
	"Ekadashi", "Dwadashi", "Trayodashi", "Chaturdashi", "Purnima",
	"Pratipada", "Dwitiya", "Tritiya", "Chaturthi", "Panchami",
	"Shashthi", "Saptami", "Ashtami", "Navami", "Dashami",
	"Ekadashi", "Dwadashi", "Trayodashi", "Chaturdashi", "Amavasya",
}

var TithiNamesHi = []string{
	"प्रतिपदा", "द्वितीया", "तृतीया", "चतुर्थी", "पंचमी",
	"षष्ठी", "सप्तमी", "अष्टमी", "नवमी", "दशमी",
	"एकादशी", "द्वादशी", "त्रयोदशी", "चतुर्दशी", "पूर्णिमा",
	"प्रतिपदा", "द्वितीया", "तृतीया", "चतुर्थी", "पंचमी",
	"षष्ठी", "सप्तमी", "अष्टमी", "नवमी", "दशमी",
	"एकादशी", "द्वादशी", "त्रयोदशी", "चतुर्दशी", "अमावस्या",
}

var TithiNamesTe = []string{
	"పాడ్యమి", "విదియ", "తదియ", "చవితి", "పంచమి",
	"షష్ఠి", "సప్తమి", "అష్టమి", "నవమి", "దశమి",
	"ఏకాదశి", "ద్వాదశి", "త్రయోదశి", "చతుర్దశి", "పౌర్ణమి",
	"పాడ్యమి", "విదియ", "తదియ", "చవితి", "పంచమి",
	"షష్ఠి", "సప్తమి", "అష్టమి", "నవమి", "దశమి",
	"ఏకాదశి", "ద్వాదశి", "త్రయోదశి", "చతుర్దశి", "అమావాస్య",
}

var NakshatraNames = []string{
	"Ashwini", "Bharani", "Krittika", "Rohini", "Mrigashira", "Ardra",
	"Punarvasu", "Pushya", "Ashlesha", "Magha", "Purva Phalguni", "Uttara Phalguni",
	"Hasta", "Chitra", "Swati", "Vishakha", "Anuradha", "Jyeshtha",
	"Mula", "Purva Ashadha", "Uttara Ashadha", "Shravana", "Dhanishta", "Shatabhisha",
	"Purva Bhadrapada", "Uttara Bhadrapada", "Revati",
}

var NakshatraNamesHi = []string{
	"अश्विनी", "भरणी", "कृत्तिका", "रोहिणी", "मृगशिरा", "आर्द्रा",
	"पुनर्वसु", "पुष्य", "आश्लेषा", "मघा", "पूर्व फाल्गुनी", "उत्तर फाल्गुनी",
	"हस्त", "चित्रा", "स्वाति", "विशाखा", "अनुराधा", "ज्येष्ठा",
	"मूल", "पूर्वाषाढ़ा", "उत्तराषाढ़ा", "श्रवण", "धनिष्ठा", "शतभिषा",
	"पूर्व भाद्रपद", "उत्तर भाद्रपद", "रेवती",
}

var YogaNames = []string{
	"Vishkambha", "Priti", "Ayushman", "Saubhagya", "Shobhana",
	"Atiganda", "Sukarma", "Dhriti", "Shula", "Ganda",
	"Vriddhi", "Dhruva", "Vyaghata", "Harshana", "Vajra",
	"Siddhi", "Vyatipata", "Variyana", "Parigha", "Shiva",
	"Siddha", "Sadhya", "Shubha", "Shukla", "Brahma",
	"Indra", "Vaidhriti",
}

var KaranaNames = []string{
	"Bava", "Balava", "Kaulava", "Taitila", "Garaja", "Vanija", "Vishti",
	"Shakuni", "Chatushpada", "Naga", "Kintughna",
}

// Rahu Kaal proportions per weekday (8ths of day duration)
// Sun=0, Mon=1, Tue=2, Wed=3, Thu=4, Fri=5, Sat=6
var rahuKaalPart = []int{8, 2, 7, 5, 6, 4, 3}
var gulikaPart = []int{7, 6, 5, 4, 3, 2, 1}
var yamagandPart = []int{5, 4, 3, 2, 1, 7, 6}

var ephemerisFlag = swe.SeflgMoseph

type Result struct {
	Date          string    `json:"date"`
	Weekday       string    `json:"weekday"`
	WeekdayHi     string    `json:"weekdayHi"`
	Tithi         Element   `json:"tithi"`
	Nakshatra     Element   `json:"nakshatra"`
	Yoga          Element   `json:"yoga"`
	Karana        Element   `json:"karana"`
	Sunrise       string    `json:"sunrise"`
	Sunset        string    `json:"sunset"`
	RahuKaal      TimeRange `json:"rahuKaal"`
	GulikaKaal    TimeRange `json:"gulikaKaal"`
	Yamagandam    TimeRange `json:"yamagandam"`
	Abhijit       TimeRange `json:"abhijit"`
	BrahmaMuhurat TimeRange `json:"brahmaMuhurat"`
}

type Element struct {
	Name   string `json:"name"`
	NameHi string `json:"nameHi"`
	NameTe string `json:"nameTe,omitempty"`
	EndsAt string `json:"endsAt,omitempty"`
}

type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Init must be called once on app start to set the ephemeris path.
func Init(ephePath string) {
	if ephePath == "" {
		ephemerisFlag = swe.SeflgMoseph
		return
	}

	swe.SetEphePath([]byte(ephePath))
	if hasSwissEphemerisFiles(ephePath) {
		ephemerisFlag = swe.SeflgSwieph
		return
	}

	// Fall back to the built-in Moshier ephemeris when packaged Swiss
	// Ephemeris data files are not present (for example in local dev).
	ephemerisFlag = swe.SeflgMoseph
}

func hasSwissEphemerisFiles(ephePath string) bool {
	planetFiles, planetErr := filepath.Glob(filepath.Join(ephePath, "sepl_*.se1"))
	moonFiles, moonErr := filepath.Glob(filepath.Join(ephePath, "semo_*.se1"))
	return planetErr == nil && moonErr == nil && len(planetFiles) > 0 && len(moonFiles) > 0
}

// Compute returns full panchang for a given date and location.
// lat / lon in degrees. tz is an IANA timezone name e.g. "Asia/Kolkata".
func Compute(date time.Time, lat, lon float64, tz string) (*Result, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	day := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, loc)

	sunriseT, err := sunRiseSet(day, lat, lon, true)
	if err != nil {
		return nil, err
	}
	sunsetT, err := sunRiseSet(day, lat, lon, false)
	if err != nil {
		return nil, err
	}

	// Compute Sun and Moon longitude at sunrise (sidereal Lahiri ayanamsa)
	swe.SetSidMode(swe.SeSidmLahiri, 0, 0)
	jd := julianDay(sunriseT.UTC())

	sunLon, err := bodyLongitude(jd, swe.SeSun)
	if err != nil {
		return nil, err
	}
	moonLon, err := bodyLongitude(jd, swe.SeMoon)
	if err != nil {
		return nil, err
	}

	// Tithi: lunar phase / 12 degrees
	diff := norm360(moonLon - sunLon)
	tithiIdx := int(diff / 12.0) // 0–29
	if tithiIdx > 29 {
		tithiIdx = 29
	}

	// Nakshatra: Moon longitude / 13.333° (= 360/27)
	nakIdx := int(moonLon / (360.0 / 27.0))
	if nakIdx > 26 {
		nakIdx = 26
	}

	// Yoga: (Sun + Moon) / 13.333°
	yogaIdx := int(norm360(sunLon+moonLon) / (360.0 / 27.0))
	if yogaIdx > 26 {
		yogaIdx = 26
	}

	// Karana: half-tithi
	karIdx := int(diff/6.0) % 11

	// Day duration in seconds for proportional periods
	dayDurMs := sunsetT.Sub(sunriseT).Milliseconds()
	partMs := dayDurMs / 8

	weekday := int(day.Weekday())
	weekdayName := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}[weekday]
	weekdayHi := []string{"रविवार", "सोमवार", "मंगलवार", "बुधवार", "गुरुवार", "शुक्रवार", "शनिवार"}[weekday]

	rahuStart := sunriseT.Add(time.Duration(int64(rahuKaalPart[weekday]-1)*partMs) * time.Millisecond)
	gulStart := sunriseT.Add(time.Duration(int64(gulikaPart[weekday]-1)*partMs) * time.Millisecond)
	yamStart := sunriseT.Add(time.Duration(int64(yamagandPart[weekday]-1)*partMs) * time.Millisecond)

	// Abhijit: midpoint ± 24 minutes
	mid := sunriseT.Add(sunsetT.Sub(sunriseT) / 2)
	abhiStart := mid.Add(-24 * time.Minute)
	abhiEnd := mid.Add(24 * time.Minute)

	// Brahma muhurat: 1.5h–48m before sunrise (next day)
	nextSunrise, err := sunRiseSet(day.AddDate(0, 0, 1), lat, lon, true)
	if err != nil {
		return nil, err
	}
	brahmaStart := nextSunrise.Add(-96 * time.Minute)
	brahmaEnd := nextSunrise.Add(-48 * time.Minute)

	return &Result{
		Date:          day.Format("2006-01-02"),
		Weekday:       weekdayName,
		WeekdayHi:     weekdayHi,
		Tithi:         Element{TithiNames[tithiIdx], TithiNamesHi[tithiIdx], TithiNamesTe[tithiIdx], ""},
		Nakshatra:     Element{NakshatraNames[nakIdx], NakshatraNamesHi[nakIdx], "", ""},
		Yoga:          Element{YogaNames[yogaIdx], "", "", ""},
		Karana:        Element{KaranaNames[karIdx], "", "", ""},
		Sunrise:       sunriseT.In(loc).Format("3:04 PM"),
		Sunset:        sunsetT.In(loc).Format("3:04 PM"),
		RahuKaal:      TimeRange{rahuStart.In(loc).Format("3:04 PM"), rahuStart.Add(time.Duration(partMs) * time.Millisecond).In(loc).Format("3:04 PM")},
		GulikaKaal:    TimeRange{gulStart.In(loc).Format("3:04 PM"), gulStart.Add(time.Duration(partMs) * time.Millisecond).In(loc).Format("3:04 PM")},
		Yamagandam:    TimeRange{yamStart.In(loc).Format("3:04 PM"), yamStart.Add(time.Duration(partMs) * time.Millisecond).In(loc).Format("3:04 PM")},
		Abhijit:       TimeRange{abhiStart.In(loc).Format("3:04 PM"), abhiEnd.In(loc).Format("3:04 PM")},
		BrahmaMuhurat: TimeRange{brahmaStart.In(loc).Format("3:04 PM"), brahmaEnd.In(loc).Format("3:04 PM")},
	}, nil
}

// ── Helpers ──

func julianDay(t time.Time) float64 {
	y, m, d := t.Year(), int(t.Month()), t.Day()
	hour := float64(t.Hour()) + float64(t.Minute())/60.0 + float64(t.Second())/3600.0
	return swe.Julday(y, m, d, hour, swe.SeGregCal)
}

func bodyLongitude(jd float64, body int) (float64, error) {
	xx := make([]float64, 6)
	serr := make([]byte, 256)
	flags := ephemerisFlag | swe.SeflgSidereal
	if ret := swe.Calc(jd, body, flags, xx, serr); ret < 0 {
		return 0, fmt.Errorf("swe_calc failed: %s", string(serr))
	}
	return xx[0], nil
}

func sunRiseSet(day time.Time, lat, lon float64, rise bool) (time.Time, error) {
	jd := julianDay(time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC))
	geopos := []float64{lon, lat, 0}
	rsmi := int32(swe.SeCalcSet)
	if rise {
		rsmi = int32(swe.SeCalcRise)
	}
	tret := make([]float64, 1)
	serr := make([]byte, 256)
	if ret := swe.RiseTrans(jd, swe.SeSun, nil, ephemerisFlag, int(rsmi), geopos, 1013.25, 15.0, tret, serr); ret < 0 {
		return time.Time{}, fmt.Errorf("rise/set failed: %s", string(serr))
	}
	// Convert JD back to UTC time
	jdRet := tret[0]
	z := jdRet + 0.5
	zInt := math.Floor(z)
	frac := z - zInt
	secs := frac * 86400
	year, mon, dy := jdToCal(int(zInt))
	t := time.Date(year, time.Month(mon), dy, 0, 0, 0, 0, time.UTC).Add(time.Duration(secs * float64(time.Second)))
	return t, nil
}

func jdToCal(jdn int) (int, int, int) {
	a := jdn + 32044
	b := (4*a + 3) / 146097
	c := a - (146097*b)/4
	d := (4*c + 3) / 1461
	e := c - (1461*d)/4
	m := (5*e + 2) / 153
	day := e - (153*m+2)/5 + 1
	month := m + 3 - 12*(m/10)
	year := 100*b + d - 4800 + m/10
	return year, month, day
}

func norm360(x float64) float64 {
	for x < 0 {
		x += 360
	}
	for x >= 360 {
		x -= 360
	}
	return x
}
