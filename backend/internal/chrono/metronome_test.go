package chrono_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/chrono"
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestPlanetaryHora(t *testing.T) {
	// Test Sunday 06:30 -> Should be Sun
	sunSunday := time.Date(2026, 9, 6, 6, 30, 0, 0, time.UTC)
	hSun := chrono.GetPlanetaryHora(sunSunday)
	if hSun.PlanetID != dasha.PlanetSun {
		t.Errorf("expected Sun hora on Sunday morning, got %s", hSun.PlanetID)
	}

	// Test Wednesday 06:30 -> Should be Mercury (Budha)
	merWednesday := time.Date(2026, 9, 9, 6, 30, 0, 0, time.UTC)
	hMer := chrono.GetPlanetaryHora(merWednesday)
	if hMer.PlanetID != dasha.PlanetMercury {
		t.Errorf("expected Mercury hora on Wednesday morning, got %s", hMer.PlanetID)
	}
}

func TestMetronomeSSEImmediatePulse(t *testing.T) {
	metro := chrono.NewMetronome(nil, 50*time.Millisecond)
	metro.Start()
	defer metro.Stop()

	req := httptest.NewRequest("GET", "/api/v1/stream/pulse", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel() // Cancel request context to exit handler
	}()

	metro.SSEHandler(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "event: pulse") {
		t.Errorf("expected SSE event: pulse, got: %s", body)
	}
	if !strings.Contains(body, `"hora"`) {
		t.Errorf("expected hora payload in pulse, got: %s", body)
	}
}
