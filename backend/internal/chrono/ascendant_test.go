package chrono_test

import (
	"math"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/chrono"
)

func TestNonLinearPlanetaryHourBoundaries(t *testing.T) {
	lat := 40.7128
	lon := -74.0060
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}

	targetDate := time.Date(2026, 9, 5, 12, 0, 0, 0, loc)
	coords := chrono.SolarCoordinates{
		Latitude:  lat,
		Longitude: lon,
	}
	schedule := chrono.CalculateDaySchedule(targetDate, coords, nil)

	if len(schedule.Hours) != 24 {
		t.Fatalf("Expected 24 hours, got %d", len(schedule.Hours))
	}

	// 1. Day strictly begins at Sunrise
	hour1 := schedule.Hours[0]
	if !hour1.StartTime.Equal(schedule.Sunrise) {
		t.Errorf("Hour 1 Start (%v) must equal Sunrise (%v)", hour1.StartTime, schedule.Sunrise)
	}

	// 2. Diurnal Hour 6 must end strictly at True Solar Noon
	hour6 := schedule.Hours[5]
	if math.Abs(hour6.EndTime.Sub(schedule.SolarNoon).Seconds()) > 1.0 {
		t.Errorf("Hour 6 End (%v) must equal Solar Noon (%v), diff: %v", hour6.EndTime, schedule.SolarNoon, hour6.EndTime.Sub(schedule.SolarNoon))
	}

	// 3. Diurnal Hour 7 must start strictly at True Solar Noon
	hour7 := schedule.Hours[6]
	if math.Abs(hour7.StartTime.Sub(schedule.SolarNoon).Seconds()) > 1.0 {
		t.Errorf("Hour 7 Start (%v) must equal Solar Noon (%v), diff: %v", hour7.StartTime, schedule.SolarNoon, hour7.StartTime.Sub(schedule.SolarNoon))
	}

	// 4. Diurnal Hour 12 must end strictly at Sunset
	hour12 := schedule.Hours[11]
	if math.Abs(hour12.EndTime.Sub(schedule.Sunset).Seconds()) > 1.0 {
		t.Errorf("Hour 12 End (%v) must equal Sunset (%v), diff: %v", hour12.EndTime, schedule.Sunset, hour12.EndTime.Sub(schedule.Sunset))
	}

	// 5. Nocturnal Hour 13 must start strictly at Sunset
	hour13 := schedule.Hours[12]
	if math.Abs(hour13.StartTime.Sub(schedule.Sunset).Seconds()) > 1.0 {
		t.Errorf("Hour 13 Start (%v) must equal Sunset (%v)", hour13.StartTime, schedule.Sunset)
	}

	// 6. Nocturnal Hour 24 must end strictly at Next Sunrise
	hour24 := schedule.Hours[23]
	if math.Abs(hour24.EndTime.Sub(schedule.NextSunrise).Seconds()) > 1.0 {
		t.Errorf("Hour 24 End (%v) must equal Next Sunrise (%v)", hour24.EndTime, schedule.NextSunrise)
	}

	// 7. Day duration != Night duration (L_day != L_night)
	dayDuration := schedule.Sunset.Sub(schedule.Sunrise)
	nightDuration := schedule.NextSunrise.Sub(schedule.Sunset)
	if dayDuration == nightDuration {
		t.Errorf("Day duration (%v) and Night duration (%v) should not be identical on Sep 5", dayDuration, nightDuration)
	}

	t.Logf("Calculated Solar Events:")
	t.Logf("  Sunrise:    %s", schedule.Sunrise.Format(time.RFC3339))
	t.Logf("  Solar Noon: %s", schedule.SolarNoon.Format(time.RFC3339))
	t.Logf("  Sunset:     %s", schedule.Sunset.Format(time.RFC3339))
	t.Logf("  Next Rise:  %s", schedule.NextSunrise.Format(time.RFC3339))
	t.Logf("  Day Length: %v | Night Length: %v", dayDuration, nightDuration)
}

func TestSubHoraObliqueAscensionPartition(t *testing.T) {
	lat := 40.7128
	lon := -74.0060
	loc, _ := time.LoadLocation("America/New_York")
	targetDate := time.Date(2026, 9, 5, 10, 30, 0, 0, loc)

	coords := chrono.SolarCoordinates{
		Latitude:  lat,
		Longitude: lon,
	}
	schedule := chrono.CalculateDaySchedule(targetDate, coords, nil)

	for _, hour := range schedule.Hours {
		if len(hour.SubHoras) != 7 {
			t.Fatalf("Hour %d expected 7 sub-horas, got %d", hour.Index, len(hour.SubHoras))
		}

		if !hour.SubHoras[0].StartTime.Equal(hour.StartTime) {
			t.Errorf("Hour %d: Sub-hora 1 start (%v) != hour start (%v)", hour.Index, hour.SubHoras[0].StartTime, hour.StartTime)
		}

		if math.Abs(hour.SubHoras[6].EndTime.Sub(hour.EndTime).Seconds()) > 0.1 {
			t.Errorf("Hour %d: Sub-hora 7 end (%v) != hour end (%v)", hour.Index, hour.SubHoras[6].EndTime, hour.EndTime)
		}

		for s := 0; s < 6; s++ {
			if !hour.SubHoras[s].EndTime.Equal(hour.SubHoras[s+1].StartTime) {
				t.Errorf("Hour %d: Sub-hora %d end (%v) != Sub-hora %d start (%v)",
					hour.Index, s+1, hour.SubHoras[s].EndTime, s+2, hour.SubHoras[s+1].StartTime)
			}
		}

		startIndex := -1
		for idx, p := range chrono.ChaldeanOrder {
			if p == hour.PlanetID {
				startIndex = idx
				break
			}
		}
		if startIndex == -1 {
			t.Fatalf("Unknown ruler %s for hour %d", hour.PlanetID, hour.Index)
		}

		for sIdx, sub := range hour.SubHoras {
			expectedPlanet := chrono.ChaldeanOrder[(startIndex+sIdx)%7]
			if sub.PlanetID != expectedPlanet {
				t.Errorf("Hour %d sub %d expected ruler %s, got %s", hour.Index, sIdx+1, expectedPlanet, sub.PlanetID)
			}
			if sub.AsuCount <= 0 {
				t.Errorf("Hour %d sub %d invalid Vedic Asu: %f", hour.Index, sIdx+1, sub.AsuCount)
			}
		}
	}
}

func TestNonLinearDurationVariance(t *testing.T) {
	lat := 51.5074
	lon := -0.1278
	loc, _ := time.LoadLocation("Europe/London")
	targetDate := time.Date(2026, 6, 21, 8, 0, 0, 0, loc)

	coords := chrono.SolarCoordinates{
		Latitude:  lat,
		Longitude: lon,
	}
	schedule := chrono.CalculateDaySchedule(targetDate, coords, nil)

	hour := schedule.Hours[0]
	durations := make([]float64, 7)
	for i, sub := range hour.SubHoras {
		durations[i] = sub.DurationMinutes
	}

	sum := 0.0
	for _, d := range durations {
		sum += d
	}
	mean := sum / 7.0

	variance := 0.0
	for _, d := range durations {
		diff := d - mean
		variance += diff * diff
	}
	variance /= 7.0

	t.Logf("London Summer Solstice Hour 1 Sub-Hora durations (minutes): %v", durations)
	t.Logf("Mean duration: %.3f mins, Variance: %.6f", mean, variance)

	if variance < 1e-6 {
		t.Errorf("Expected non-linear sub-hora durations with non-zero variance, got variance %.8f", variance)
	}
}

func TestLocalApparentSolarTimeAndEquationOfTime(t *testing.T) {
	coords := chrono.SolarCoordinates{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	solar := chrono.CalculateSolarTimes(time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC), coords)
	lastAtNoon := chrono.ConvertToLocalApparentSolarTime(solar.SolarNoon, coords.Longitude)

	t.Logf("True Solar Noon UTC: %s, LAST at Solar Noon: %s", solar.SolarNoon.Format(time.RFC3339), lastAtNoon.Format("15:04:05"))
	diffFrom12 := math.Abs(float64(lastAtNoon.Hour()*60+lastAtNoon.Minute()) - 12.0*60.0)
	if diffFrom12 > 1.0 {
		t.Errorf("Expected LAST at Solar Noon to be ~12:00, got %s (diff: %f mins)", lastAtNoon.Format("15:04:05"), diffFrom12)
	}
}
