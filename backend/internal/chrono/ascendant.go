package chrono

import (
	"math"
	"time"
)

// CalculateSiderealTime computes Greenwich Mean Sidereal Time (GMST) and Local Sidereal Time (LST) in degrees.
func CalculateSiderealTime(t time.Time, lon float64) (gmstDeg, lstDeg float64) {
	y := t.Year()
	m := int(t.Month())
	d := float64(t.Day()) + (float64(t.Hour()) / 24.0) + (float64(t.Minute()) / 1440.0) + (float64(t.Second()) / 86400.0) + (float64(t.Nanosecond()) / 86400e9)

	if m <= 2 {
		y -= 1
		m += 12
	}
	a := math.Floor(float64(y) / 100.0)
	b := 2.0 - a + math.Floor(a/4.0)
	jd := math.Floor(365.25*float64(y+4716)) + math.Floor(30.6001*float64(m+1)) + d + b - 1524.5

	// Julian centuries from J2000.0
	tCent := (jd - 2451545.0) / 36525.0

	// GMST in degrees (Meeus formula)
	gmst := 280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*tCent*tCent - (tCent*tCent*tCent)/38710000.0
	gmst = math.Mod(gmst, 360.0)
	if gmst < 0 {
		gmst += 360.0
	}

	// LST = GMST + Longitude (East positive)
	lst := math.Mod(gmst+lon, 360.0)
	if lst < 0 {
		lst += 360.0
	}

	return gmst, lst
}

// CalculateAscendant computes the instantaneous ecliptic longitude of the Ascendant (Lagna) in degrees [0, 360).
func CalculateAscendant(t time.Time, lat, lon float64) float64 {
	_, lst := CalculateSiderealTime(t, lon)

	// Obliquity of the Ecliptic eps
	y := t.Year()
	m := int(t.Month())
	d := float64(t.Day())
	if m <= 2 {
		y -= 1
		m += 12
	}
	a := math.Floor(float64(y) / 100.0)
	b := 2.0 - a + math.Floor(a/4.0)
	jd := math.Floor(365.25*float64(y+4716)) + math.Floor(30.6001*float64(m+1)) + d + b - 1524.5
	tCent := (jd - 2451545.0) / 36525.0
	eps := 23.439291 - 0.0130042*tCent

	lstRad := lst * deg2rad
	latRad := lat * deg2rad
	epsRad := eps * deg2rad

	// Standard Ascendant spherical formula:
	// y = -cos(LST)
	// x = sin(LST)*cos(eps) + tan(lat)*sin(eps)
	yComp := -math.Cos(lstRad)
	xComp := math.Sin(lstRad)*math.Cos(epsRad) + math.Tan(latRad)*math.Sin(epsRad)

	ascRad := math.Atan2(yComp, xComp)
	ascDeg := ascRad * rad2deg

	if ascDeg < 0 {
		ascDeg += 360.0
	}
	return math.Mod(ascDeg, 360.0)
}

// SubHoraPartition represents one of the 7 non-linear sub-divisions of a Planetary Hour.
type SubHoraPartition struct {
	Index               int       `json:"index"` // 1 to 7
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	DurationMinutes     float64   `json:"duration_minutes"`
	LagnaArcStart       float64   `json:"lagna_arc_start"`       // Degree of Ascendant at start
	LagnaArcEnd         float64   `json:"lagna_arc_end"`         // Degree of Ascendant at end
	LagnaArcTraversed   float64   `json:"lagna_arc_traversed"`   // Arc degrees traversed
	AsuCount            float64   `json:"asu_count"`             // Vedic Asu duration (1 Asu = 4 seconds)
}

// PartitionHoraByObliqueAscension calculates the 7 non-linear sub-hora intervals based on equal Ascendant (Lagna) arc traversal.
func PartitionHoraByObliqueAscension(startTime, endTime time.Time, lat, lon float64) []SubHoraPartition {
	totalDuration := endTime.Sub(startTime)
	if totalDuration <= 0 {
		return nil
	}

	lagnaStart := CalculateAscendant(startTime, lat, lon)
	lagnaEnd := CalculateAscendant(endTime, lat, lon)

	totalArc := lagnaEnd - lagnaStart
	if totalArc < 0 {
		totalArc += 360.0
	}
	// If arc is negligible or extreme polar anomaly, distribute linearly
	if totalArc < 0.1 {
		totalArc = 360.0 / 24.0 // nominal ~15 degrees
	}

	subArc := totalArc / 7.0

	partitions := make([]SubHoraPartition, 7)
	currTime := startTime
	currLagna := lagnaStart

	for i := 0; i < 7; i++ {
		targetLagna := math.Mod(lagnaStart+float64(i+1)*subArc, 360.0)
		var nextTime time.Time

		if i == 6 {
			// Exact boundary anchor at parent hour end
			nextTime = endTime
			targetLagna = lagnaEnd
		} else {
			// Find time within [startTime, endTime] where Lagna == targetLagna
			nextTime = findTimeForLagna(currTime, endTime, targetLagna, lat, lon, lagnaStart, subArc*float64(i+1))
		}

		dur := nextTime.Sub(currTime)
		durMins := dur.Minutes()
		durSecs := dur.Seconds()
		asu := durSecs / 4.0 // 1 Asu = 4 seconds of time (1' of celestial ascension)

		arcTraversed := subArc
		if i == 6 {
			arcTraversed = targetLagna - currLagna
			if arcTraversed < 0 {
				arcTraversed += 360.0
			}
		}

		partitions[i] = SubHoraPartition{
			Index:             i + 1,
			StartTime:         currTime,
			EndTime:           nextTime,
			DurationMinutes:   math.Round(durMins*100) / 100,
			LagnaArcStart:     math.Round(currLagna*100) / 100,
			LagnaArcEnd:       math.Round(targetLagna*100) / 100,
			LagnaArcTraversed: math.Round(arcTraversed*100) / 100,
			AsuCount:          math.Round(asu*10) / 10,
		}

		currTime = nextTime
		currLagna = targetLagna
	}

	return partitions
}

// findTimeForLagna finds timestamp t in [tMin, tMax] matching target cumulative arc from lagnaRef.
func findTimeForLagna(tMin, tMax time.Time, targetLagna, lat, lon, lagnaRef, targetCumulativeArc float64) time.Time {
	low := tMin
	high := tMax

	// Binary search / bisection root solver (18 iterations achieves sub-second accuracy)
	for iter := 0; iter < 18; iter++ {
		mid := low.Add(high.Sub(low) / 2)
		midLagna := CalculateAscendant(mid, lat, lon)

		arc := midLagna - lagnaRef
		if arc < 0 {
			arc += 360.0
		}

		if arc < targetCumulativeArc {
			low = mid
		} else {
			high = mid
		}
	}

	return low.Add(high.Sub(low) / 2)
}
