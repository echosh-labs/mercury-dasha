package dasha_test

import (
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestSacredMetalsCompleteness(t *testing.T) {
	if len(dasha.SacredMetals) != 7 {
		t.Fatalf("expected 7 sacred metals, got %d", len(dasha.SacredMetals))
	}

	// Verify Quicksilver corresponds to Mercury
	qs, ok := dasha.MetalsByPlanet[dasha.PlanetMercury]
	if !ok {
		t.Fatalf("expected Quicksilver for PlanetMercury")
	}
	if qs.ID != "quicksilver" || qs.Symbol != "☿" {
		t.Errorf("unexpected quicksilver properties: %+v", qs)
	}

	// Verify Gold corresponds to Sun
	gold, ok := dasha.MetalsByPlanet[dasha.PlanetSun]
	if !ok || gold.ID != "gold" || gold.Symbol != "☉" {
		t.Errorf("unexpected gold properties: %+v", gold)
	}
}

func TestHermeticAxiomsCompleteness(t *testing.T) {
	if len(dasha.HermeticAxioms) != 7 {
		t.Fatalf("expected 7 Hermetic axioms, got %d", len(dasha.HermeticAxioms))
	}

	// Axiom 1 should be Mentalism
	a1 := dasha.HermeticAxioms[0]
	if a1.Number != 1 || a1.GoverningPlanet != dasha.PlanetMercury {
		t.Errorf("expected Axiom 1 Mentalism with Mercury, got %+v", a1)
	}
}

func TestResolveAlchemicalAlignment(t *testing.T) {
	align := dasha.ResolveAlchemicalAlignment(dasha.PlanetMercury)
	if align.ActiveMetal.ID != "quicksilver" {
		t.Errorf("expected quicksilver, got %s", align.ActiveMetal.ID)
	}
	if align.GoverningAxiom.Title != "The Principle of Mentalism" {
		t.Errorf("expected Mentalism, got %s", align.GoverningAxiom.Title)
	}
	if align.MagnumOpusStage.Name != "Separation" && align.MagnumOpusStage.Name != "Distillation" {
		t.Errorf("expected Separation or Distillation for Mercury, got %s", align.MagnumOpusStage.Name)
	}
}

func TestEngineDomainMethods(t *testing.T) {
	engine := dasha.NewEngine(nil)

	// Overview
	overview := engine.GetOverview()
	if overview["mahadasha_lord"] != "Mercury (17-Year Cycle)" {
		t.Errorf("expected Mercury overview, got %v", overview["mahadasha_lord"])
	}

	// Nakshatras
	naks := engine.GetNakshatras("")
	if naks["total"] != 27 {
		t.Errorf("expected 27 total nakshatras, got %v", naks["total"])
	}

	// Nakshatras filtered by mercury
	mercNaks := engine.GetNakshatras("mercury")
	if mercNaks["count"] != 3 {
		t.Errorf("expected 3 mercury nakshatras, got %v", mercNaks["count"])
	}

	// Alchemy
	alchem := engine.GetAlchemy(dasha.PlanetMercury)
	if alchem["sovereign_element"] == "" {
		t.Errorf("expected sovereign_element in alchemy response")
	}
}
