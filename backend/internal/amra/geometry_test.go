package amra

import (
	"math"
	"strings"
	"testing"
)

func TestCalculateKairiCurve(t *testing.T) {
	params := DefaultKairiParams()
	result := CalculateKairiCurve(params)

	if len(result.Points) != params.Steps+1 {
		t.Fatalf("expected %d points, got %d", params.Steps+1, len(result.Points))
	}

	if !strings.HasPrefix(result.SVGPath, "M ") || !strings.HasSuffix(result.SVGPath, " Z") {
		t.Errorf("SVG path malformed: %s", result.SVGPath)
	}

	// Verify loop closure: t = -pi and t = pi should evaluate to nearly identical points
	firstPoint := result.Points[0]
	lastPoint := result.Points[len(result.Points)-1]
	dist := math.Hypot(firstPoint.X-lastPoint.X, firstPoint.Y-lastPoint.Y)
	if dist > 0.05 {
		t.Errorf("curve did not close cleanly: distance between start and end is %f", dist)
	}

	// Verify bounding box consistency
	if result.BoundingBox.MinX >= result.BoundingBox.MaxX {
		t.Errorf("invalid X bounding box: min=%f, max=%f", result.BoundingBox.MinX, result.BoundingBox.MaxX)
	}
	if result.BoundingBox.MinY >= result.BoundingBox.MaxY {
		t.Errorf("invalid Y bounding box: min=%f, max=%f", result.BoundingBox.MinY, result.BoundingBox.MaxY)
	}

	// Arc length should be positive and realistic for a ~300-pixel height shape
	if result.ArcLength < 200 || result.ArcLength > 3000 {
		t.Errorf("unexpected arc length: %f", result.ArcLength)
	}
}

func TestCalculateTransmutation(t *testing.T) {
	// Test 1: High solar fire, moderate shadow -> Pakva state
	pakva := CalculateTransmutation(0.3, 0.9)
	if pakva.State != "pakva" {
		t.Errorf("expected state 'pakva', got '%s'", pakva.State)
	}
	if pakva.DevicRatio <= pakva.AsuricRatio {
		t.Errorf("expected Devic ratio > Asuric ratio in pakva state")
	}
	if len(pakva.Demons) != 6 {
		t.Fatalf("expected 6 Arishadvarga vectors, got %d", len(pakva.Demons))
	}
	for _, d := range pakva.Demons {
		if d.Teaching == "" {
			t.Errorf("demon %s has empty teaching", d.Name)
		}
		if d.ShadowDistortion == "" {
			t.Errorf("demon %s has empty shadow distortion", d.Name)
		}
	}

	// Test 2: Low solar fire -> Āma state
	ama := CalculateTransmutation(0.8, 0.2)
	if ama.State != "ama" {
		t.Errorf("expected state 'ama', got '%s'", ama.State)
	}

	// Test 3: Moderate solar fire and shadow -> Manthan (churning) state
	manthan := CalculateTransmutation(0.5, 0.5)
	if manthan.State != "manthan" {
		t.Errorf("expected state 'manthan', got '%s'", manthan.State)
	}
}

func TestCalculateMangoLeaves(t *testing.T) {
	origin := Point2D{X: 0, Y: -120}
	leaves := CalculateMangoLeaves(origin)

	if len(leaves) != 5 {
		t.Fatalf("expected 5 sacred leaves, got %d", len(leaves))
	}

	for i, leaf := range leaves {
		if leaf.Length <= 0 {
			t.Errorf("leaf %d has non-positive length: %f", i, leaf.Length)
		}
		if !strings.HasPrefix(leaf.SVGPath, "M ") || !strings.HasSuffix(leaf.SVGPath, " Z") {
			t.Errorf("leaf %d has invalid SVG path: %s", i, leaf.SVGPath)
		}
	}
}

func TestGenerateAmraGeometry(t *testing.T) {
	params := DefaultKairiParams()
	geo := GenerateAmraGeometry(params, 0.4, 0.8)

	if len(geo.Body.Points) == 0 {
		t.Errorf("body points should not be empty")
	}
	if len(geo.Bija.Points) == 0 {
		t.Errorf("bija points should not be empty")
	}
	if len(geo.Leaves) != 5 {
		t.Errorf("expected 5 leaves, got %d", len(geo.Leaves))
	}
	if len(geo.MathematicalFormulas) == 0 {
		t.Errorf("mathematical formulas dictionary should not be empty")
	}
}
