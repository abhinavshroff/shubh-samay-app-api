package panchang

import (
	"fmt"
	"sort"
	"strings"
	"time"

	swe "github.com/mshafiee/swephgo"
)

const (
	MuhuratActivityTravel   = "travel"
	MuhuratActivityMarriage = "marriage"
	MuhuratActivityBusiness = "business"
	MuhuratActivityHome     = "home"
	MuhuratActivityVehicle  = "vehicle"
)

const (
	muhuratGradeExcellent  = "Excellent"
	muhuratGradeGood       = "Good"
	muhuratGradeAuspicious = "Auspicious"
)

var blockedTithiNumbers = map[int]bool{4: true, 8: true, 9: true, 14: true, 30: true}
var blockedYogaNames = map[string]bool{
	"Vishkambha": true,
	"Atiganda":   true,
	"Vyaghata":   true,
	"Vajra":      true,
	"Vyatipata":  true,
	"Parigha":    true,
	"Vaidhriti":  true,
}
var blockedWeekdays = map[time.Weekday]bool{time.Tuesday: true, time.Saturday: true}
var marriageMonths = map[int]bool{
	MonthChaitra:    true,
	MonthVaishakha:  true,
	MonthJyeshtha:   true,
	MonthBhadrapada: true,
	MonthAshwin:     true,
	MonthKartika:    true,
	MonthMagha:      true,
}

type MuhuratRequest struct {
	Activity string
	From     time.Time
	Days     int
	Limit    int
	Lat      float64
	Lon      float64
	Timezone string
}

type MuhuratResult struct {
	Activity     string          `json:"activity"`
	Date         string          `json:"date"`
	DisplayDate  string          `json:"displayDate"`
	Weekday      string          `json:"weekday"`
	Time         TimeRange       `json:"time"`
	Score        int             `json:"score"`
	Grade        string          `json:"grade"`
	Stars        string          `json:"stars"`
	Highlighted  bool            `json:"highlighted"`
	BorderColor  string          `json:"borderColor,omitempty"`
	Meta         string          `json:"meta"`
	Factors      MuhuratFactors  `json:"factors"`
	Explanations []string        `json:"explanations"`
	Details      map[string]bool `json:"details,omitempty"`
}

type MuhuratFactors struct {
	Nakshatra MuhuratFactor  `json:"nakshatra"`
	Tithi     MuhuratFactor  `json:"tithi"`
	Weekday   MuhuratFactor  `json:"weekday"`
	Yoga      MuhuratFactor  `json:"yoga"`
	Month     *MuhuratFactor `json:"month,omitempty"`
	Kharmas   *MuhuratFactor `json:"kharmas,omitempty"`
}

type MuhuratFactor struct {
	Name      string `json:"name"`
	Number    int    `json:"number,omitempty"`
	Passed    bool   `json:"passed"`
	Preferred bool   `json:"preferred,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type muhuratActivityRule struct {
	activity            string
	label               string
	searchDays          int
	nakshatras          map[string]bool
	preferredTithis     map[int]bool
	allowedTithis       map[int]bool
	preferredWeekdays   map[time.Weekday]bool
	allowedWeekdays     map[time.Weekday]bool
	requiresVivahMonths bool
}

type muhuratDaySnapshot struct {
	day        time.Time
	sunrise    time.Time
	sunset     time.Time
	sunLon     float64
	moonLon    float64
	tithiIndex int
	tithiName  string
	yogaIndex  int
	yogaName   string
	nakIndex   int
	nakName    string
}

var muhuratRules = map[string]muhuratActivityRule{
	MuhuratActivityTravel: {
		activity:          MuhuratActivityTravel,
		label:             "Travel",
		searchDays:        30,
		nakshatras:        nameSet("Ashwini", "Mrigashira", "Ardra", "Punarvasu", "Pushya", "Hasta", "Anuradha", "Shravana", "Dhanishta", "Revati"),
		preferredTithis:   intSet(2, 3, 5, 7, 10, 11, 13),
		allowedTithis:     intSet(1, 2, 3, 5, 6, 7, 10, 11, 12, 13, 15),
		preferredWeekdays: weekdaySet(time.Wednesday, time.Thursday, time.Friday, time.Monday),
		allowedWeekdays:   weekdaySet(time.Sunday, time.Monday, time.Wednesday, time.Thursday, time.Friday),
	},
	MuhuratActivityMarriage: {
		activity:            MuhuratActivityMarriage,
		label:               "Marriage",
		searchDays:          90,
		nakshatras:          nameSet("Rohini", "Mrigashira", "Magha", "Uttara Phalguni", "Hasta", "Swati", "Anuradha", "Mula", "Uttara Ashadha", "Uttara Bhadrapada", "Revati"),
		preferredTithis:     intSet(2, 3, 5, 7, 10, 11, 13),
		allowedTithis:       intSet(2, 3, 5, 6, 7, 10, 11, 12, 13),
		preferredWeekdays:   weekdaySet(time.Wednesday, time.Thursday, time.Friday, time.Monday),
		allowedWeekdays:     weekdaySet(time.Monday, time.Wednesday, time.Thursday, time.Friday, time.Sunday),
		requiresVivahMonths: true,
	},
	MuhuratActivityBusiness: {
		activity:          MuhuratActivityBusiness,
		label:             "Business",
		searchDays:        30,
		nakshatras:        nameSet("Ashwini", "Rohini", "Mrigashira", "Pushya", "Uttara Phalguni", "Hasta", "Chitra", "Swati", "Anuradha", "Uttara Ashadha", "Shravana", "Dhanishta", "Revati"),
		preferredTithis:   intSet(2, 3, 5, 7, 10, 11, 13),
		allowedTithis:     intSet(1, 2, 3, 5, 6, 7, 10, 11, 12, 13, 15),
		preferredWeekdays: weekdaySet(time.Wednesday, time.Thursday, time.Friday, time.Monday),
		allowedWeekdays:   weekdaySet(time.Sunday, time.Monday, time.Wednesday, time.Thursday, time.Friday),
	},
	MuhuratActivityHome: {
		activity:          MuhuratActivityHome,
		label:             "Home",
		searchDays:        30,
		nakshatras:        nameSet("Rohini", "Mrigashira", "Uttara Phalguni", "Hasta", "Anuradha", "Uttara Ashadha", "Shravana", "Dhanishta", "Uttara Bhadrapada", "Revati"),
		preferredTithis:   intSet(2, 3, 5, 7, 10, 11, 13),
		allowedTithis:     intSet(1, 2, 3, 5, 6, 7, 10, 11, 12, 13, 15),
		preferredWeekdays: weekdaySet(time.Wednesday, time.Thursday, time.Friday, time.Monday),
		allowedWeekdays:   weekdaySet(time.Sunday, time.Monday, time.Wednesday, time.Thursday, time.Friday),
	},
	MuhuratActivityVehicle: {
		activity:          MuhuratActivityVehicle,
		label:             "Vehicle",
		searchDays:        30,
		nakshatras:        nameSet("Ashwini", "Rohini", "Mrigashira", "Punarvasu", "Pushya", "Hasta", "Chitra", "Swati", "Anuradha", "Shravana", "Dhanishta", "Revati"),
		preferredTithis:   intSet(2, 3, 5, 7, 10, 11, 13),
		allowedTithis:     intSet(1, 2, 3, 5, 6, 7, 10, 11, 12, 13, 15),
		preferredWeekdays: weekdaySet(time.Wednesday, time.Thursday, time.Friday, time.Monday),
		allowedWeekdays:   weekdaySet(time.Sunday, time.Monday, time.Wednesday, time.Thursday, time.Friday),
	},
}

func FindMuhurat(req MuhuratRequest) ([]MuhuratResult, error) {
	rule, ok := muhuratRules[NormalizeMuhuratActivity(req.Activity)]
	if !ok {
		return nil, fmt.Errorf("unsupported muhurat activity %q", req.Activity)
	}
	loc, err := time.LoadLocation(req.Timezone)
	if err != nil {
		return nil, err
	}
	days := req.Days
	if days <= 0 {
		days = rule.searchDays
	}
	if days > 180 {
		days = 180
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}
	if limit > 25 {
		limit = 25
	}

	from := req.From
	if from.IsZero() {
		from = time.Now()
	}
	start := time.Date(from.In(loc).Year(), from.In(loc).Month(), from.In(loc).Day(), 12, 0, 0, 0, loc)
	results := make([]MuhuratResult, 0, limit)
	for offset := 0; offset < days; offset++ {
		snapshot, err := muhuratSnapshot(start.AddDate(0, 0, offset), req.Lat, req.Lon)
		if err != nil {
			return nil, err
		}
		result, ok, err := evaluateMuhurat(rule, snapshot, loc)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		result.Activity = rule.activity
		results = append(results, result)
		if len(results) == limit {
			break
		}
	}
	if len(results) > 0 {
		results[0].Highlighted = true
		results[0].BorderColor = "#FF9933"
	}
	return results, nil
}

func NormalizeMuhuratActivity(activity string) string {
	switch strings.ToLower(strings.TrimSpace(activity)) {
	case "", "trip", "journey":
		return MuhuratActivityTravel
	case "vivah", "wedding":
		return MuhuratActivityMarriage
	case "business", "work", "office", "shop":
		return MuhuratActivityBusiness
	case "home", "house", "griha", "griha-pravesh", "griha_pravesh", "housewarming":
		return MuhuratActivityHome
	case "vehicle", "car", "bike":
		return MuhuratActivityVehicle
	default:
		return strings.ToLower(strings.TrimSpace(activity))
	}
}

func SupportedMuhuratActivities() []string {
	activities := make([]string, 0, len(muhuratRules))
	for activity := range muhuratRules {
		activities = append(activities, activity)
	}
	sort.Strings(activities)
	return activities
}

func muhuratSnapshot(day time.Time, lat, lon float64) (muhuratDaySnapshot, error) {
	sunrise, err := sunRiseSet(day, lat, lon, true)
	if err != nil {
		return muhuratDaySnapshot{}, err
	}
	sunset, err := sunRiseSet(day, lat, lon, false)
	if err != nil {
		return muhuratDaySnapshot{}, err
	}
	swe.SetSidMode(swe.SeSidmLahiri, 0, 0)
	jd := julianDay(sunrise.UTC())
	sunLon, err := bodyLongitude(jd, swe.SeSun)
	if err != nil {
		return muhuratDaySnapshot{}, err
	}
	moonLon, err := bodyLongitude(jd, swe.SeMoon)
	if err != nil {
		return muhuratDaySnapshot{}, err
	}
	diff := norm360(moonLon - sunLon)
	tithiIdx := int(diff / 12.0)
	if tithiIdx > 29 {
		tithiIdx = 29
	}
	nakIdx := int(moonLon / (360.0 / 27.0))
	if nakIdx > 26 {
		nakIdx = 26
	}
	yogaIdx := int(norm360(sunLon+moonLon) / (360.0 / 27.0))
	if yogaIdx > 26 {
		yogaIdx = 26
	}
	return muhuratDaySnapshot{day: day, sunrise: sunrise, sunset: sunset, sunLon: sunLon, moonLon: moonLon, tithiIndex: tithiIdx, tithiName: TithiNames[tithiIdx], yogaIndex: yogaIdx, yogaName: YogaNames[yogaIdx], nakIndex: nakIdx, nakName: NakshatraNames[nakIdx]}, nil
}

func evaluateMuhurat(rule muhuratActivityRule, snapshot muhuratDaySnapshot, loc *time.Location) (MuhuratResult, bool, error) {
	weekday := snapshot.day.In(loc).Weekday()
	tithiNumber := muhuratTithiNumber(snapshot.tithiIndex)
	tithiPakshaNumber := snapshot.tithiIndex%15 + 1
	factors := MuhuratFactors{
		Nakshatra: MuhuratFactor{Name: snapshot.nakName, Number: snapshot.nakIndex + 1, Passed: rule.nakshatras[snapshot.nakName], Reason: "Moon nakshatra checked for activity"},
		Tithi:     MuhuratFactor{Name: snapshot.tithiName, Number: tithiNumber, Passed: !blockedMuhuratTithi(snapshot.tithiIndex) && rule.allowedTithis[tithiPakshaNumber], Preferred: rule.preferredTithis[tithiPakshaNumber], Reason: "Tithi checked at local sunrise"},
		Weekday:   MuhuratFactor{Name: weekday.String(), Number: int(weekday), Passed: !blockedWeekdays[weekday] && rule.allowedWeekdays[weekday], Preferred: rule.preferredWeekdays[weekday], Reason: "Tuesday and Saturday are always excluded"},
		Yoga:      MuhuratFactor{Name: snapshot.yogaName, Number: snapshot.yogaIndex + 1, Passed: !blockedYogaNames[snapshot.yogaName], Reason: "Sun + Moon yoga checked at local sunrise"},
	}
	if !factors.Nakshatra.Passed || !factors.Tithi.Passed || !factors.Weekday.Passed || !factors.Yoga.Passed {
		return MuhuratResult{}, false, nil
	}

	if rule.requiresVivahMonths {
		sunSign := int(snapshot.sunLon / 30.0)
		kharmas := sunSign == 8 || sunSign == 11 // sidereal Sagittarius or Pisces
		factors.Kharmas = &MuhuratFactor{Name: sunSignName(sunSign), Number: sunSign + 1, Passed: !kharmas, Reason: "Marriage excludes Kharmas: Sun in Sagittarius or Pisces"}
		month, err := lunarMonthAt(snapshot.sunrise, CalendarNorth)
		if err != nil {
			return MuhuratResult{}, false, err
		}
		factors.Month = &MuhuratFactor{Name: lunarMonthNames[month], Number: month + 1, Passed: marriageMonths[month], Reason: "Marriage is restricted to traditional Vivah lunar months"}
		if kharmas || !factors.Month.Passed {
			return MuhuratResult{}, false, nil
		}
	}

	score := muhuratScore(factors)
	grade, stars := muhuratGrade(score)
	start, end := muhuratWindow(snapshot, loc)
	metaParts := []string{snapshot.nakName + " nakshatra", snapshot.yogaName + " yoga", snapshot.tithiName + " tithi"}
	if rule.requiresVivahMonths {
		metaParts = append(metaParts, factors.Month.Name+" month")
	}

	return MuhuratResult{
		Date:        snapshot.day.In(loc).Format("2006-01-02"),
		DisplayDate: snapshot.day.In(loc).Format("Mon, 2 Jan"),
		Weekday:     weekday.String(),
		Time:        TimeRange{From: start.Format("3:04 PM"), To: end.Format("3:04 PM")},
		Score:       score,
		Grade:       grade,
		Stars:       stars,
		Meta:        strings.Join(metaParts, " • "),
		Factors:     factors,
		Explanations: []string{
			"All results pass the independent nakshatra, tithi, weekday, and yoga checks.",
			"Chaturthi, Ashtami, Navami, Chaturdashi, Amavasya, Tuesday, Saturday, and inauspicious yogas are excluded.",
		},
		Details: map[string]bool{
			"nakshatra": true,
			"tithi":     true,
			"weekday":   true,
			"yoga":      true,
		},
	}, true, nil
}

func muhuratTithiNumber(index int) int {
	if index == 29 {
		return 30
	}
	return index%15 + 1
}

func blockedMuhuratTithi(index int) bool {
	return blockedTithiNumbers[muhuratTithiNumber(index)]
}

func muhuratScore(factors MuhuratFactors) int {
	score := 60
	if factors.Nakshatra.Passed {
		score += 15
	}
	if factors.Tithi.Preferred {
		score += 10
	} else if factors.Tithi.Passed {
		score += 5
	}
	if factors.Weekday.Preferred {
		score += 10
	} else if factors.Weekday.Passed {
		score += 5
	}
	if factors.Yoga.Passed && !blockedYogaNames[factors.Yoga.Name] {
		score += 5
	}
	if factors.Month != nil && factors.Month.Passed {
		score += 5
	}
	if score > 100 {
		return 100
	}
	return score
}

func muhuratGrade(score int) (string, string) {
	if score >= 90 {
		return muhuratGradeExcellent, "★★★"
	}
	if score >= 78 {
		return muhuratGradeGood, "★★"
	}
	return muhuratGradeAuspicious, "★"
}

func muhuratWindow(snapshot muhuratDaySnapshot, loc *time.Location) (time.Time, time.Time) {
	start := snapshot.sunrise.In(loc).Add(48 * time.Minute)
	end := start.Add(90 * time.Minute)
	latestEnd := snapshot.sunset.In(loc).Add(-48 * time.Minute)
	if end.After(latestEnd) {
		end = latestEnd
	}
	return start, end
}

func sunSignName(sign int) string {
	names := []string{"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo", "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"}
	if sign < 0 || sign >= len(names) {
		return "Unknown"
	}
	return names[sign]
}

func nameSet(names ...string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

func intSet(values ...int) map[int]bool {
	set := make(map[int]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func weekdaySet(values ...time.Weekday) map[time.Weekday]bool {
	set := make(map[time.Weekday]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
