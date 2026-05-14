package panchang

import (
	"fmt"
	"sort"
	"time"

	swe "github.com/mshafiee/swephgo"
)

const (
	MonthChaitra = iota
	MonthVaishakha
	MonthJyeshtha
	MonthAshadha
	MonthShravana
	MonthBhadrapada
	MonthAshwin
	MonthKartika
	MonthMargashirsha
	MonthPausha
	MonthMagha
	MonthPhalguna
)

var lunarMonthNames = []string{
	"Chaitra", "Vaishakha", "Jyeshtha", "Ashadha", "Shravana", "Bhadrapada",
	"Ashwin", "Kartika", "Margashirsha", "Pausha", "Magha", "Phalguna",
}

var lunarMonthNamesHi = []string{
	"चैत्र", "वैशाख", "ज्येष्ठ", "आषाढ़", "श्रावण", "भाद्रपद",
	"आश्विन", "कार्तिक", "मार्गशीर्ष", "पौष", "माघ", "फाल्गुन",
}

// Festival is a calculated Hindu festival occurrence returned by /v1/festivals.
type Festival struct {
	Date         string `json:"date"`
	ISODate      string `json:"isoDate"`
	Name         string `json:"name"`
	NameHi       string `json:"nameHi,omitempty"`
	NameTe       string `json:"nameTe,omitempty"`
	Tithi        string `json:"tithi"`
	TithiHi      string `json:"tithiHi"`
	LunarMonth   string `json:"lunarMonth"`
	LunarMonthHi string `json:"lunarMonthHi"`
	Region       string `json:"region,omitempty"`
	Significance string `json:"significance,omitempty"`
	Rule         string `json:"rule"`
	DaysAway     int    `json:"daysAway"`
	ObservedAt   string `json:"observedAt,omitempty"`
}

type festivalRule struct {
	name         string
	nameHi       string
	nameTe       string
	month        int
	tithiIndex   int
	observe      festivalObserve
	region       string
	significance string
}

type festivalObserve int

const (
	observeSunrise festivalObserve = iota
	observeNoon
	observePradosh
	observeMidnight
	observeMoonrise
)

var festivalRules = []festivalRule{
	{"Maha Shivaratri", "महाशिवरात्रि", "", MonthPhalguna, 28, observeMidnight, "national", "Night of Shiva observed on Krishna Chaturdashi."},
	{"Holi", "होली", "", MonthPhalguna, 14, observeSunrise, "national", "Festival of colours on Phalguna Purnima."},
	{"Rama Navami", "राम नवमी", "", MonthChaitra, 8, observeNoon, "national", "Birth of Lord Rama on Chaitra Shukla Navami."},
	{"Hanuman Jayanti", "हनुमान जयंती", "", MonthChaitra, 14, observeSunrise, "national", "Observed on Chaitra Purnima in many traditions."},
	{"Akshaya Tritiya", "अक्षय तृतीया", "", MonthVaishakha, 2, observeSunrise, "national", "Vaishakha Shukla Tritiya."},
	{"Buddha Purnima", "बुद्ध पूर्णिमा", "", MonthVaishakha, 14, observeSunrise, "national", "Vaishakha Purnima."},
	{"Guru Purnima", "गुरु पूर्णिमा", "", MonthAshadha, 14, observeSunrise, "national", "Ashadha Purnima dedicated to gurus."},
	{"Nag Panchami", "नाग पंचमी", "", MonthShravana, 4, observeSunrise, "national", "Shravana Shukla Panchami."},
	{"Raksha Bandhan", "रक्षा बंधन", "", MonthShravana, 14, observeSunrise, "national", "Shravana Purnima."},
	{"Krishna Janmashtami", "कृष्ण जन्माष्टमी", "", MonthBhadrapada, 22, observeMidnight, "national", "Bhadrapada Krishna Ashtami at midnight."},
	{"Ganesh Chaturthi", "गणेश चतुर्थी", "", MonthBhadrapada, 3, observeNoon, "national", "Bhadrapada Shukla Chaturthi."},
	{"Sharad Navaratri Begins", "शारदीय नवरात्रि प्रारंभ", "", MonthAshwin, 0, observeSunrise, "national", "Ashwin Shukla Pratipada."},
	{"Dussehra", "दशहरा", "", MonthAshwin, 9, observeNoon, "national", "Ashwin Shukla Dashami."},
	{"Karwa Chauth", "करवा चौथ", "", MonthKartika, 18, observeMoonrise, "north", "Kartika Krishna Chaturthi at moonrise."},
	{"Dhanteras", "धनतेरस", "", MonthKartika, 27, observePradosh, "national", "Kartika Krishna Trayodashi during pradosh kaal."},
	{"Diwali", "दीपावली", "", MonthKartika, 29, observePradosh, "national", "Kartika Amavasya during pradosh kaal."},
	{"Govardhan Puja", "गोवर्धन पूजा", "", MonthKartika, 0, observeSunrise, "national", "Kartika Shukla Pratipada."},
	{"Bhai Dooj", "भाई दूज", "", MonthKartika, 1, observeSunrise, "national", "Kartika Shukla Dwitiya."},
	{"Dev Diwali", "देव दीपावली", "", MonthKartika, 14, observeSunrise, "national", "Kartika Purnima."},
	{"Gita Jayanti", "गीता जयंती", "", MonthMargashirsha, 10, observeSunrise, "national", "Margashirsha Shukla Ekadashi."},
	{"Vasant Panchami", "वसंत पंचमी", "", MonthMagha, 4, observeSunrise, "national", "Magha Shukla Panchami."},
}

// Festivals computes Hindu festival dates from their fixed lunar rules over the
// requested date window. Lunar month names are derived from the Moon longitude
// at the next full moon, not from the current Sun sign.
func Festivals(from time.Time, days int, lat, lon float64, tz, calendar string) ([]Festival, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	if days <= 0 {
		days = 120
	}
	if days > 730 {
		days = 730
	}

	start := time.Date(from.In(loc).Year(), from.In(loc).Month(), from.In(loc).Day(), 12, 0, 0, 0, loc)
	items := make([]Festival, 0, len(festivalRules))
	seen := map[string]bool{}
	for offset := 0; offset < days; offset++ {
		day := start.AddDate(0, 0, offset)
		sunrise, err := sunRiseSet(day, lat, lon, true)
		if err != nil {
			return nil, err
		}
		sunset, err := sunRiseSet(day, lat, lon, false)
		if err != nil {
			return nil, err
		}
		for _, rule := range festivalRules {
			observedAt := festivalObservationTime(rule.observe, day, sunrise, sunset, lat, lon, loc)
			info, err := tithiInfoAt(observedAt)
			if err != nil {
				return nil, err
			}
			if info.Index != rule.tithiIndex {
				continue
			}
			// Festival rules are stored in the purnimanta month convention used by
			// the common North Indian festival names, so rule matching stays stable
			// even when the caller uses an amanta regional calendar elsewhere.
			month, err := lunarMonthAt(observedAt, CalendarNorth)
			if err != nil {
				return nil, err
			}
			if month != rule.month {
				continue
			}
			isoDate := day.In(loc).Format("2006-01-02")
			key := isoDate + ":" + rule.name
			if seen[key] {
				continue
			}
			seen[key] = true
			items = append(items, newFestival(day, offset, rule, info, observedAt, loc))
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].ISODate == items[j].ISODate {
			return items[i].Name < items[j].Name
		}
		return items[i].ISODate < items[j].ISODate
	})
	return items, nil
}

func festivalObservationTime(observe festivalObserve, day, sunrise, sunset time.Time, lat, lon float64, loc *time.Location) time.Time {
	switch observe {
	case observeNoon:
		return time.Date(day.In(loc).Year(), day.In(loc).Month(), day.In(loc).Day(), 12, 0, 0, 0, loc)
	case observePradosh:
		return sunset.Add(48 * time.Minute)
	case observeMidnight:
		return time.Date(day.In(loc).Year(), day.In(loc).Month(), day.In(loc).Day(), 0, 0, 0, 0, loc)
	case observeMoonrise:
		moonrise, err := riseSet(day, lat, lon, swe.SeMoon, true)
		if err == nil && moonrise.In(loc).Day() == day.In(loc).Day() {
			return moonrise
		}
		return sunset.Add(2 * time.Hour)
	default:
		return sunrise
	}
}

func newFestival(day time.Time, daysAway int, rule festivalRule, info tithiInfo, observedAt time.Time, loc *time.Location) Festival {
	monthName := lunarMonthNames[rule.month]
	monthNameHi := lunarMonthNamesHi[rule.month]
	return Festival{
		Date:         day.In(loc).Format("2 Jan"),
		ISODate:      day.In(loc).Format("2006-01-02"),
		Name:         rule.name,
		NameHi:       rule.nameHi,
		NameTe:       rule.nameTe,
		Tithi:        fmt.Sprintf("%s %s", monthName, info.Paksha+" "+info.Name),
		TithiHi:      fmt.Sprintf("%s %s %s", monthNameHi, info.PakshaHi, info.NameHi),
		LunarMonth:   monthName,
		LunarMonthHi: monthNameHi,
		Region:       rule.region,
		Significance: rule.significance,
		Rule:         rule.significance,
		DaysAway:     daysAway,
		ObservedAt:   observedAt.In(loc).Format("3:04 PM"),
	}
}

func lunarMonthAt(t time.Time, calendar string) (int, error) {
	cal := NormalizeCalendar(calendar)
	fullMoonTime, err := fullMoonForLunarMonth(t, cal)
	if err != nil {
		return 0, err
	}
	jd := julianDay(fullMoonTime.UTC())
	moonLon, err := bodyLongitude(jd, swe.SeMoon)
	if err != nil {
		return 0, err
	}
	return lunarMonthFromMoonLongitude(moonLon), nil
}

func fullMoonForLunarMonth(t time.Time, calendar string) (time.Time, error) {
	if calendar == CalendarSouth {
		info, err := tithiInfoAt(t)
		if err != nil {
			return time.Time{}, err
		}
		if info.Index >= 15 {
			return previousFullMoon(t)
		}
	}
	return nextFullMoon(t)
}

func lunarMonthFromMoonLongitude(moonLon float64) int {
	moonSign := int(norm360(moonLon) / 30.0)
	if moonSign > 11 {
		moonSign = 11
	}
	// Lunar months are named for the nakshatra/rashi region occupied by the
	// Moon at full moon: Chaitra full moon is in Virgo, Vaishakha in Libra, ...,
	// Kartika in Aries. This sign mapping is direct Moon-longitude based.
	return (moonSign + 7) % 12
}

func nextFullMoon(from time.Time) (time.Time, error) {
	return phaseCrossing(from, 180, true)
}

func previousFullMoon(from time.Time) (time.Time, error) {
	return phaseCrossing(from, 180, false)
}

func phaseCrossing(from time.Time, target float64, forward bool) (time.Time, error) {
	step := 6 * time.Hour
	if !forward {
		step = -step
	}
	prevTime := from
	prevPhase, err := lunarPhase(prevTime)
	if err != nil {
		return time.Time{}, err
	}
	prevDelta := signedPhaseDelta(prevPhase, target, forward)
	for i := 0; i < 140; i++ {
		nextTime := prevTime.Add(step)
		nextPhase, err := lunarPhase(nextTime)
		if err != nil {
			return time.Time{}, err
		}
		nextDelta := signedPhaseDelta(nextPhase, target, forward)
		if nextDelta > prevDelta || nextDelta == 0 {
			lo, hi := prevTime, nextTime
			if hi.Before(lo) {
				lo, hi = hi, lo
			}
			return refinePhaseCrossing(lo, hi, target)
		}
		prevTime, prevDelta = nextTime, nextDelta
	}
	return time.Time{}, fmt.Errorf("full moon phase crossing not found near %s", from.Format(time.RFC3339))
}

func refinePhaseCrossing(lo, hi time.Time, target float64) (time.Time, error) {
	base, err := lunarPhase(lo)
	if err != nil {
		return time.Time{}, err
	}
	for hi.Sub(lo) > time.Minute {
		mid := lo.Add(hi.Sub(lo) / 2)
		phase, err := lunarPhase(mid)
		if err != nil {
			return time.Time{}, err
		}
		if norm360(phase-base) < norm360(target-base) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi, nil
}

func signedPhaseDelta(phase, target float64, forward bool) float64 {
	if forward {
		return norm360(target - phase)
	}
	return norm360(phase - target)
}

func lunarPhase(t time.Time) (float64, error) {
	swe.SetSidMode(swe.SeSidmLahiri, 0, 0)
	jd := julianDay(t.UTC())
	sunLon, err := bodyLongitude(jd, swe.SeSun)
	if err != nil {
		return 0, err
	}
	moonLon, err := bodyLongitude(jd, swe.SeMoon)
	if err != nil {
		return 0, err
	}
	return norm360(moonLon - sunLon), nil
}
