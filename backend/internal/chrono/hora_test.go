package chrono_test

import (
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/chrono"
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestSolarAstronomicalCalculations(t *testing.T) {
	// Equinox date: March 20, 2026 at Greenwich Equator (0, 0)
	equinox := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	coords := chrono.SolarCoordinates{Latitude: 0.0, Longitude: 0.0}

	solar := chrono.CalculateSolarTimes(equinox, coords)

	if solar.Sunrise.IsZero() || solar.Sunset.IsZero() {
		t.Fatalf("expected valid sunrise and sunset, got zero time")
	}

	if solar.Sunrise.After(solar.Sunset) {
		t.Errorf("sunrise %v cannot be after sunset %v", solar.Sunrise, solar.Sunset)
	}

	dayHours := solar.DayDuration.Hours()
	if dayHours < 11.5 || dayHours > 12.5 {
		t.Errorf("expected ~12h daylight on equinox at equator, got %f", dayHours)
	}
}

func Test24PlanetaryHourChaldeanCycle(t *testing.T) {
	// Wednesday, Sept 9, 2026
	testTime := time.Date(2026, 9, 9, 14, 0, 0, 0, time.UTC)
	coords := chrono.SolarCoordinates{Latitude: 40.7128, Longitude: -74.0060}

	sched := chrono.CalculateDaySchedule(testTime, coords, nil)

	if len(sched.Hours) != 24 {
		t.Fatalf("expected exactly 24 planetary hours, got %d", len(sched.Hours))
	}

	// 1. Day Lord for Wednesday must be Mercury (Budha)
	if sched.DayLordPlanetID != dasha.PlanetMercury {
		t.Errorf("expected Wednesday Day Lord to be Mercury, got %s", sched.DayLordPlanetID)
	}

	// 2. First hour (Index 1) must be Mercury
	if sched.Hours[0].PlanetID != dasha.PlanetMercury {
		t.Errorf("expected Hour 1 on Wednesday to be Mercury, got %s", sched.Hours[0].PlanetID)
	}

	// 3. Verify Chaldean descending sequence across all 24 hours:
	// Saturn -> Jupiter -> Mars -> Sun -> Venus -> Mercury -> Moon
	for i := 0; i < 23; i++ {
		currentPlanet := sched.Hours[i].PlanetID
		nextPlanet := sched.Hours[i+1].PlanetID

		// Find current index in ChaldeanOrder
		cIdx := -1
		for idx, p := range chrono.ChaldeanOrder {
			if p == currentPlanet {
				cIdx = idx
				break
			}
		}
		expectedNext := chrono.ChaldeanOrder[(cIdx+1)%7]
		if nextPlanet != expectedNext {
			t.Errorf("hour %d (%s) -> hour %d expected %s, got %s",
				i+1, currentPlanet, i+2, expectedNext, nextPlanet)
		}
	}

	// 4. Mathematical proof: 25th hour must be the Lord of the next day!
	// Wednesday -> next day is Thursday (Jupiter)
	lastHourPlanet := sched.Hours[23].PlanetID
	lastIdx := -1
	for idx, p := range chrono.ChaldeanOrder {
		if p == lastHourPlanet {
			lastIdx = idx
			break
		}
	}
	hour25Planet := chrono.ChaldeanOrder[(lastIdx+1)%7]
	if hour25Planet != dasha.PlanetJupiter {
		t.Errorf("expected 25th hour to be Thursday Lord Jupiter, got %s", hour25Planet)
	}

	// 5. Diurnal / Nocturnal demarcation
	for i := 0; i < 12; i++ {
		if !sched.Hours[i].Diurnal {
			t.Errorf("hour %d should be diurnal", i+1)
		}
	}
	for i := 12; i < 24; i++ {
		if sched.Hours[i].Diurnal {
			t.Errorf("hour %d should be nocturnal", i+1)
		}
	}
}

func TestDailyLivingGuidanceCompleteness(t *testing.T) {
	classicalPlanets := []dasha.PlanetID{
		dasha.PlanetSun,
		dasha.PlanetMoon,
		dasha.PlanetMars,
		dasha.PlanetMercury,
		dasha.PlanetJupiter,
		dasha.PlanetVenus,
		dasha.PlanetSaturn,
	}

	for _, p := range classicalPlanets {
		guidance, exists := chrono.PlanetaryHoraGuidances[p]
		if !exists {
			t.Fatalf("missing guidance for planet %s", p)
		}
		if len(guidance.SuitableActivities) == 0 {
			t.Errorf("no suitable activities for %s", p)
		}
		if len(guidance.UnsuitableActivities) == 0 {
			t.Errorf("no unsuitable activities for %s", p)
		}
		if guidance.BriefApplication == "" {
			t.Errorf("empty brief application for %s", p)
		}
	}
}

func TestSymbioticDashaResonance(t *testing.T) {
	refTime := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	coords := chrono.SolarCoordinates{Latitude: 40.7128, Longitude: -74.0060}

	// Construct mock Dasha profile in Mercury Mahadasha and Venus Antardasha
	profile := &dasha.DashaProfile{
		ID:   "test-profile",
		Name: "Test Seeker",
		ActiveSnapshot: dasha.ActivePeriodSnapshot{
			Mahadasha: dasha.DashaPeriod{
				Planet:     dasha.PlanetMercury,
				PlanetName: "Mercury",
			},
			Antardasha: dasha.DashaPeriod{
				Planet:     dasha.PlanetVenus,
				PlanetName: "Venus",
			},
			Pratyantardasha: dasha.DashaPeriod{
				Planet:     dasha.PlanetSun,
				PlanetName: "Sun",
			},
		},
	}

	sched := chrono.CalculateDaySchedule(refTime, coords, profile)

	if sched.DashaResonance == nil {
		t.Fatalf("expected DashaResonance to be computed when profile provided")
	}

	// Verify resonance classification
	activePlanet := sched.ActiveHour.PlanetID
	switch activePlanet {
	case dasha.PlanetMercury:
		if sched.DashaResonance.ResonanceType != "sovereign" {
			t.Errorf("expected sovereign resonance for Mercury Hora in Mercury Mahadasha, got %s", sched.DashaResonance.ResonanceType)
		}
	case dasha.PlanetVenus:
		if sched.DashaResonance.ResonanceType != "catalytic" {
			t.Errorf("expected catalytic resonance for Venus Hora in Venus Antardasha, got %s", sched.DashaResonance.ResonanceType)
		}
	case dasha.PlanetSun:
		if sched.DashaResonance.ResonanceType != "immediate" {
			t.Errorf("expected immediate resonance for Sun Hora in Sun Pratyantardasha, got %s", sched.DashaResonance.ResonanceType)
		}
	}
}
