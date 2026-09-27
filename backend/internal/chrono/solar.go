package chrono

import (
	"math"
	"time"
)

const (
	deg2rad = math.Pi / 180.0
	rad2deg = 180.0 / math.Pi
	// Standard zenith for topocentric apparent sunrise/sunset:
	// 90° geometric + 34' atmospheric refraction + 16' solar semi-diameter = 90°50' = 90.833333°
	solarZenith = 90.83333333333333
)

// SolarCoordinates represents geographic location.
type SolarCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SolarDayTimes captures astronomical solar events and exact non-linear temporal phase durations.
type SolarDayTimes struct {
	Date                  string        `json:"date"`
	Sunrise               time.Time     `json:"sunrise"`
	SolarNoon             time.Time     `json:"solar_noon"`
	Sunset                time.Time     `json:"sunset"`
	NextSunrise           time.Time     `json:"next_sunrise"`
	SolarMidnight         time.Time     `json:"solar_midnight"`
	EquationOfTimeMinutes float64       `json:"equation_of_time_minutes"`
	DayDuration           time.Duration `json:"day_duration"`
	NightDuration         time.Duration `json:"night_duration"`
	MorningDuration       time.Duration `json:"morning_duration"`
	AfternoonDuration     time.Duration `json:"afternoon_duration"`
	MorningHourDuration   time.Duration `json:"morning_hour_duration"`   // L_morning = (SolarNoon - Sunrise)/6
	AfternoonHourDuration time.Duration `json:"afternoon_hour_duration"` // L_afternoon = (Sunset - SolarNoon)/6
	NightHourDuration     time.Duration `json:"night_hour_duration"`     // L_night = (NextSunrise - Sunset)/12
}

// DefaultCoordinates provides the sovereign baseline reference point (St. Catharines / Niagara: 43.1594° N, -79.2469° W).
var DefaultCoordinates = SolarCoordinates{
	Latitude:  43.1594,
	Longitude: -79.2469,
}

// CalculateSolarTimes calculates Sunrise, Sunset, Solar Noon, and Next Sunrise for a given date and coordinates.
// When coords are (0,0), it uses the equatorial prime meridian reference where solar time naturally synchronizes with UTC.
func CalculateSolarTimes(t time.Time, coords SolarCoordinates) SolarDayTimes {
	year, month, day := t.Date()
	dateMidnight := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	dateStr := dateMidnight.Format("2006-01-02")

	sunrise, noon, sunset, eot := computeSingleDaySolar(dateMidnight, coords.Latitude, coords.Longitude)

	// Next day's sunrise for nocturnal hora calculations
	nextMidnight := dateMidnight.AddDate(0, 0, 1)
	nextSunrise, nextNoon, _, _ := computeSingleDaySolar(nextMidnight, coords.Latitude, coords.Longitude)

	dayDur := sunset.Sub(sunrise)
	nightDur := nextSunrise.Sub(sunset)
	morningDur := noon.Sub(sunrise)
	afternoonDur := sunset.Sub(noon)

	// True solar midnight between sunset and next sunrise:
	// Anti-meridian passage is exactly halfway between today's solar noon and tomorrow's solar noon:
	solarMidnight := noon.Add(nextNoon.Sub(noon) / 2)

	morningHourDur := morningDur / 6
	afternoonHourDur := afternoonDur / 6
	nightHourDur := nightDur / 12

	return SolarDayTimes{
		Date:                  dateStr,
		Sunrise:               sunrise,
		SolarNoon:             noon,
		Sunset:                sunset,
		NextSunrise:           nextSunrise,
		SolarMidnight:         solarMidnight,
		EquationOfTimeMinutes: math.Round(eot*100) / 100,
		DayDuration:           dayDur,
		NightDuration:         nightDur,
		MorningDuration:       morningDur,
		AfternoonDuration:     afternoonDur,
		MorningHourDuration:   morningHourDur,
		AfternoonHourDuration: afternoonHourDur,
		NightHourDuration:     nightHourDur,
	}
}

// ComputeEquationOfTime computes the Equation of Time (in minutes) at any given UTC instant.
func ComputeEquationOfTime(t time.Time) float64 {
	year, month, day := t.Date()
	midnight := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	_, _, _, eot := computeSingleDaySolar(midnight, 0, 0)
	return eot
}

// ConvertToLocalApparentSolarTime converts a UTC timestamp to Local Apparent Solar Time (LAST):
// LAST = UTC_Time + (4 * Longitude_Minutes) + Equation_Of_Time
func ConvertToLocalApparentSolarTime(t time.Time, lon float64) time.Time {
	eot := ComputeEquationOfTime(t)
	// Longitude offset: 4 minutes per degree of longitude (East positive)
	lonOffset := time.Duration(lon * 4.0 * float64(time.Minute))
	eotOffset := time.Duration(eot * float64(time.Minute))

	return t.Add(lonOffset).Add(eotOffset)
}

func computeSingleDaySolar(midnightUTC time.Time, lat, lon float64) (sunrise, noon, sunset time.Time, eot float64) {
	// Julian day at 0h UT
	y := midnightUTC.Year()
	m := int(midnightUTC.Month())
	d := float64(midnightUTC.Day())

	if m <= 2 {
		y -= 1
		m += 12
	}
	A := math.Floor(float64(y) / 100.0)
	B := 2.0 - A + math.Floor(A/4.0)
	jd := math.Floor(365.25*float64(y+4716)) + math.Floor(30.6001*float64(m+1)) + d + B - 1524.5

	T := (jd - 2451545.0) / 36525.0

	// Mean Solar Longitude (deg)
	L0 := math.Mod(280.46646+T*(36000.76983+T*0.0003032), 360.0)
	if L0 < 0 {
		L0 += 360.0
	}

	// Mean Solar Anomaly (deg)
	M := 357.52911 + T*(35999.05029-0.0001537*T)
	Mrad := M * deg2rad

	// Eccentricity
	e := 0.016708634 - T*(0.000042037+0.0000001267*T)

	// Equation of Center (deg)
	C := math.Sin(Mrad)*(1.914602-T*(0.004817+0.000014*T)) +
		math.Sin(2.0*Mrad)*(0.019993-0.000101*T) +
		math.Sin(3.0*Mrad)*0.000289

	// Sun's True Longitude (deg)
	sunTrueLong := L0 + C

	// Sun's Apparent Longitude (deg)
	omega := 125.04 - 1934.136*T
	lambda := sunTrueLong - 0.00569 - 0.00478*math.Sin(omega*deg2rad)
	lambdaRad := lambda * deg2rad

	// Obliquity of the Ecliptic (deg)
	eps0 := 23.0 + (26.0+(21.448-T*(46.815+T*(0.00059-T*0.001813)))/60.0)/60.0
	eps := eps0 + 0.00256*math.Cos(omega*deg2rad)
	epsRad := eps * deg2rad

	// Sun Declination (rad)
	sinDec := math.Sin(epsRad) * math.Sin(lambdaRad)
	dec := math.Asin(sinDec)

	// Equation of Time (minutes)
	tanHalfEps := math.Tan(epsRad / 2.0)
	yVar := tanHalfEps * tanHalfEps
	L0Rad := L0 * deg2rad

	eot = 4.0 * rad2deg * (yVar*math.Sin(2.0*L0Rad) -
		2.0*e*math.Sin(Mrad) +
		4.0*e*yVar*math.Sin(Mrad)*math.Cos(2.0*L0Rad) -
		0.5*yVar*yVar*math.Sin(4.0*L0Rad) -
		1.25*e*e*math.Sin(2.0*Mrad))

	// Solar Noon in minutes from midnight UTC
	solarNoonMinutes := 720.0 - (4.0 * lon) - eot
	noon = midnightUTC.Add(time.Duration(solarNoonMinutes * float64(time.Minute)))

	// Hour Angle for Sunrise/Sunset
	latRad := lat * deg2rad
	cosHA := (math.Cos(solarZenith*deg2rad) - math.Sin(latRad)*sinDec) / (math.Cos(latRad) * math.Cos(dec))

	var haMinutes float64
	if cosHA >= 1.0 {
		// Polar night: sun never rises, fallback to noon +/- 1 minute
		haMinutes = 0
	} else if cosHA <= -1.0 {
		// Midnight sun: sun never sets, 24h sunlight
		haMinutes = 720.0
	} else {
		haDeg := math.Acos(cosHA) * rad2deg
		haMinutes = haDeg * 4.0 // 4 minutes per degree
	}

	sunrise = noon.Add(-time.Duration(haMinutes * float64(time.Minute)))
	sunset = noon.Add(time.Duration(haMinutes * float64(time.Minute)))

	return sunrise, noon, sunset, eot
}
