package panchang

import (
	"fmt"
	"strings"
	"time"

	swe "github.com/mshafiee/swephgo"
)

const (
	CalendarNorth = "north"
	CalendarSouth = "south"
)

var pakshaNames = []string{"Shukla", "Krishna"}
var pakshaNamesHi = []string{"शुक्ल", "कृष्ण"}

// LunarDay is a calculated vrata / lunar observance returned by /v1/lunar-days.
type LunarDay struct {
	Date       string `json:"date"`
	Type       string `json:"type"`
	Label      string `json:"label"`
	LabelHi    string `json:"labelHi,omitempty"`
	Tithi      string `json:"tithi"`
	TithiHi    string `json:"tithiHi"`
	Paksha     string `json:"paksha"`
	PakshaHi   string `json:"pakshaHi"`
	Calendar   string `json:"calendar"`
	Rule       string `json:"rule"`
	DaysAway   int    `json:"daysAway"`
	StartsAt   string `json:"startsAt,omitempty"`
	EndsAt     string `json:"endsAt,omitempty"`
	ObservedAt string `json:"observedAt,omitempty"`
}

// NormalizeCalendar maps user settings to one of the supported regional calendars.
// North uses the purnimanta convention; South uses the amanta convention.
func NormalizeCalendar(calendar string) string {
	switch strings.ToLower(strings.TrimSpace(calendar)) {
	case CalendarSouth, "southern", "south-indian", "south_indian", "amanta", "amant":
		return CalendarSouth
	case CalendarNorth, "northern", "north-indian", "north_indian", "purnimanta", "purnimant", "":
		return CalendarNorth
	default:
		return CalendarNorth
	}
}

// LunarDays returns real calculated Ekadashi, Purnima, Amavasya, Pradosh, and
// Sankashti dates for the requested location and regional calendar.
func LunarDays(from time.Time, days int, lat, lon float64, tz, calendar string) ([]LunarDay, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	if days <= 0 {
		days = 45
	}
	if days > 370 {
		days = 370
	}

	cal := NormalizeCalendar(calendar)
	start := time.Date(from.In(loc).Year(), from.In(loc).Month(), from.In(loc).Day(), 12, 0, 0, 0, loc)
	items := make([]LunarDay, 0, days/3)
	for offset := 0; offset < days; offset++ {
		day := start.AddDate(0, 0, offset)
		dayItems, err := lunarDaysForDate(day, offset, lat, lon, loc, cal)
		if err != nil {
			return nil, err
		}
		items = append(items, dayItems...)
	}
	return items, nil
}

func lunarDaysForDate(day time.Time, daysAway int, lat, lon float64, loc *time.Location, calendar string) ([]LunarDay, error) {
	sunrise, err := sunRiseSet(day, lat, lon, true)
	if err != nil {
		return nil, err
	}
	sunset, err := sunRiseSet(day, lat, lon, false)
	if err != nil {
		return nil, err
	}

	sunriseInfo, err := tithiInfoAt(sunrise)
	if err != nil {
		return nil, err
	}
	endsAt, _ := nextTithiBoundary(sunrise, sunriseInfo.Index)

	items := make([]LunarDay, 0, 2)
	sunriseRule := "active at local sunrise"
	switch sunriseInfo.Index {
	case 10, 25:
		items = append(items, newLunarDay(day, daysAway, "ekadashi", "Ekadashi", "एकादशी", sunriseInfo, calendar, sunriseRule, sunrise, time.Time{}, endsAt, loc))
	case 14:
		items = append(items, newLunarDay(day, daysAway, "purnima", "Purnima", "पूर्णिमा", sunriseInfo, calendar, sunriseRule, sunrise, time.Time{}, endsAt, loc))
	case 29:
		items = append(items, newLunarDay(day, daysAway, "amavasya", "Amavasya", "अमावस्या", sunriseInfo, calendar, sunriseRule, sunrise, time.Time{}, endsAt, loc))
	}

	pradoshAt := sunset.Add(48 * time.Minute)
	pradoshInfo, err := tithiInfoAt(pradoshAt)
	if err != nil {
		return nil, err
	}
	if pradoshInfo.Index == 12 || pradoshInfo.Index == 27 {
		items = append(items, newLunarDay(day, daysAway, "pradosh", "Pradosh", "प्रदोष", pradoshInfo, calendar, "Trayodashi active during pradosh kaal", pradoshAt, time.Time{}, time.Time{}, loc))
	}

	moonrise, err := riseSet(day, lat, lon, swe.SeMoon, true)
	observedAt := moonrise
	if err != nil || moonrise.In(loc).Day() != day.In(loc).Day() {
		observedAt = sunset.Add(2 * time.Hour)
	}
	sankashtiInfo, err := tithiInfoAt(observedAt)
	if err != nil {
		return nil, err
	}
	if sankashtiInfo.Index == 18 {
		items = append(items, newLunarDay(day, daysAway, "sankashti", "Sankashti Chaturthi", "संकष्टी चतुर्थी", sankashtiInfo, calendar, "Krishna Chaturthi active at moonrise", observedAt, time.Time{}, time.Time{}, loc))
	}

	return items, nil
}

type tithiInfo struct {
	Index    int
	Number   int
	Name     string
	NameHi   string
	Paksha   string
	PakshaHi string
}

func tithiInfoAt(t time.Time) (tithiInfo, error) {
	swe.SetSidMode(swe.SeSidmLahiri, 0, 0)
	jd := julianDay(t.UTC())
	sunLon, err := bodyLongitude(jd, swe.SeSun)
	if err != nil {
		return tithiInfo{}, err
	}
	moonLon, err := bodyLongitude(jd, swe.SeMoon)
	if err != nil {
		return tithiInfo{}, err
	}
	idx := int(norm360(moonLon-sunLon) / 12.0)
	if idx > 29 {
		idx = 29
	}
	pakshaIdx := 0
	if idx >= 15 {
		pakshaIdx = 1
	}
	return tithiInfo{Index: idx, Number: idx + 1, Name: TithiNames[idx], NameHi: TithiNamesHi[idx], Paksha: pakshaNames[pakshaIdx], PakshaHi: pakshaNamesHi[pakshaIdx]}, nil
}

func nextTithiBoundary(from time.Time, currentIdx int) (time.Time, error) {
	lo := from
	hi := from.Add(30 * time.Hour)
	for hi.Sub(lo) > time.Minute {
		mid := lo.Add(hi.Sub(lo) / 2)
		info, err := tithiInfoAt(mid)
		if err != nil {
			return time.Time{}, err
		}
		if info.Index == currentIdx {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi, nil
}

func newLunarDay(day time.Time, daysAway int, kind, label, labelHi string, info tithiInfo, calendar, rule string, observedAt, startsAt, endsAt time.Time, loc *time.Location) LunarDay {
	if calendar == CalendarNorth && info.Paksha == "Krishna" {
		rule += " (North Indian purnimanta calendar)"
	} else if calendar == CalendarSouth {
		rule += " (South Indian amanta calendar)"
	}

	item := LunarDay{
		Date:     day.In(loc).Format("2006-01-02"),
		Type:     kind,
		Label:    label,
		LabelHi:  labelHi,
		Tithi:    fmt.Sprintf("%s %s", info.Paksha, info.Name),
		TithiHi:  fmt.Sprintf("%s %s", info.PakshaHi, info.NameHi),
		Paksha:   info.Paksha,
		PakshaHi: info.PakshaHi,
		Calendar: calendar,
		Rule:     rule,
		DaysAway: daysAway,
	}
	if !observedAt.IsZero() {
		item.ObservedAt = observedAt.In(loc).Format("3:04 PM")
	}
	if !startsAt.IsZero() {
		item.StartsAt = startsAt.In(loc).Format("3:04 PM")
	}
	if !endsAt.IsZero() {
		item.EndsAt = endsAt.In(loc).Format("3:04 PM")
	}
	return item
}

func riseSet(day time.Time, lat, lon float64, body int, rise bool) (time.Time, error) {
	jd := julianDay(time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC))
	geopos := []float64{lon, lat, 0}
	rsmi := int32(swe.SeCalcSet)
	if rise {
		rsmi = int32(swe.SeCalcRise)
	}
	tret := make([]float64, 1)
	serr := make([]byte, 256)
	if ret := swe.RiseTrans(jd, body, nil, ephemerisFlag, int(rsmi), geopos, 1013.25, 15.0, tret, serr); ret < 0 {
		return time.Time{}, fmt.Errorf("rise/set failed: %s", string(serr))
	}
	return jdFloatToTime(tret[0]), nil
}
