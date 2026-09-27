package amra

import (
	"fmt"
	"math"
	"strings"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// GoldenRatioPhi is the divine proportion governing sacred harmonic structures.
const GoldenRatioPhi = 1.618033988749895

// Point2D represents a Cartesian 2D coordinate.
type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BoundingBox holds the min and max coordinates of a generated curve.
type BoundingBox struct {
	MinX float64 `json:"min_x"`
	MaxX float64 `json:"max_x"`
	MinY float64 `json:"min_y"`
	MaxY float64 `json:"max_y"`
}

// KairiParams defines the mathematical coefficients of the parametric Mango/Kairi curve:
//
//	x(t) = Scale * (Rx * sin(t) * (1 + Alpha * cos(t)) + Gamma * sin^3(t/2))
//	y(t) = Scale * (-Ry * cos(t) + Ry * Beta * sin^2(t/2) * cos(t) + OffsetY)
//
// for parameter t ∈ [-π, π].
type KairiParams struct {
	Rx      float64 `json:"rx"`       // Semi-major horizontal radius (Garbha lateral fullness)
	Ry      float64 `json:"ry"`       // Semi-major vertical radius
	Alpha   float64 `json:"alpha"`    // Asymmetry coefficient (belly swell on receptive side)
	Beta    float64 `json:"beta"`     // Egg inflection coefficient (Hiranyagarbha egg lift)
	Gamma   float64 `json:"gamma"`    // Crest hook deflection (Pratyahara inward recurve)
	Steps   int     `json:"steps"`    // Numerical resolution for curve discretization
	Scale   float64 `json:"scale"`    // Global spatial multiplier
	OffsetY float64 `json:"offset_y"` // Vertical center displacement
}

// DefaultKairiParams returns sacred canonical proportions for the Āmra silhouette.
func DefaultKairiParams() KairiParams {
	return KairiParams{
		Rx:      125.0,
		Ry:      175.0,
		Alpha:   0.35,
		Beta:    0.22,
		Gamma:   25.0,
		Steps:   120,
		Scale:   1.0,
		OffsetY: 30.0,
	}
}

// KairiCurveResult encapsulates the computed curve geometry and SVG path.
type KairiCurveResult struct {
	Points         []Point2D   `json:"points"`
	SVGPath        string      `json:"svg_path"`
	BoundingBox    BoundingBox `json:"bounding_box"`
	ArcLength      float64     `json:"arc_length"`
	GoldenRatioPhi float64     `json:"golden_ratio_phi"`
}

// CalculateKairiCurve computes the parametric points and SVG path for given parameters.
func CalculateKairiCurve(params KairiParams) KairiCurveResult {
	if params.Steps < 10 {
		params.Steps = 60
	}
	if params.Scale <= 0 {
		params.Scale = 1.0
	}

	points := make([]Point2D, params.Steps+1)
	var pathBuilder strings.Builder

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64
	var totalArcLength float64

	for i := 0; i <= params.Steps; i++ {
		// Parameter t traverses [-π, π]
		t := (float64(i) / float64(params.Steps)) * 2.0 * math.Pi - math.Pi

		sinT := math.Sin(t)
		cosT := math.Cos(t)

		// Base teardrop with lateral belly asymmetry (Alpha)
		baseX := params.Rx * sinT * (1.0 + params.Alpha*cosT)

		// Crest hook with Pratyāhāra recurve:
		// (1 + cos(t))^2 vanishes identically at t = ±π, ensuring exact C1 closure at the base,
		// while modulating the apex (t=0) with the S-curve inflection.
		crestEnvelope := math.Pow(1.0+cosT, 2) / 4.0 // Normalized to [0, 1] at apex
		hookX := params.Gamma * crestEnvelope * (1.0 - 0.5*sinT)

		x := params.Scale * (baseX + hookX)
		y := params.Scale * (-params.Ry*cosT + (params.Ry*params.Beta)*math.Pow(math.Sin(t/2.0), 2)*cosT + params.OffsetY)

		p := Point2D{
			X: math.Round(x*100) / 100,
			Y: math.Round(y*100) / 100,
		}
		points[i] = p

		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}

		if i == 0 {
			pathBuilder.WriteString(fmt.Sprintf("M %.2f %.2f", p.X, p.Y))
		} else {
			pathBuilder.WriteString(fmt.Sprintf(" L %.2f %.2f", p.X, p.Y))
			dx := points[i].X - points[i-1].X
			dy := points[i].Y - points[i-1].Y
			totalArcLength += math.Hypot(dx, dy)
		}
	}
	pathBuilder.WriteString(" Z")

	return KairiCurveResult{
		Points:  points,
		SVGPath: pathBuilder.String(),
		BoundingBox: BoundingBox{
			MinX: minX,
			MaxX: maxX,
			MinY: minY,
			MaxY: maxY,
		},
		ArcLength:      math.Round(totalArcLength*100) / 100,
		GoldenRatioPhi: GoldenRatioPhi,
	}
}

// ArishadvargaVector represents one of the six classical inner adversaries (demons)
// acting as rotational torque around the consciousness center in Samudra Manthan.
type ArishadvargaVector struct {
	Index            int     `json:"index"`
	Key              string  `json:"key"`
	Name             string  `json:"name"`              // Sanskrit designation
	Meaning          string  `json:"meaning"`           // English translation
	ShadowDistortion string  `json:"shadow_distortion"` // Shadow distortion
	AngleRad         float64 `json:"angle_rad"`         // Radial bearing
	Distance         float64 `json:"distance"`          // Spatial excursion
	Coord            Point2D `json:"coord"`             // Cartesian coordinate
	TransmutedAs     string  `json:"transmuted_as"`     // Converted dharmic virtue
	Teaching         string  `json:"teaching"`          // Full transformative guidance
}

// TransmutationState details the dynamic balance between unprocessed psychic poison (āma)
// and solar awareness (surya agni), yielding the immortal nectar (amṛta).
type TransmutationState struct {
	State             string               `json:"state"`              // "ama" (raw), "manthan" (churning), "pakva" (ripe)
	Description       string               `json:"description"`
	ShadowTension     float64              `json:"shadow_tension"`     // 0.0 to 1.0 (magnitude of unprocessed shadow)
	SolarFire         float64              `json:"solar_fire"`         // 0.0 to 1.0 (intensity of conscious awareness)
	DevicRatio        float64              `json:"devic_ratio"`        // Percentage light / conscious alignment
	AsuricRatio       float64              `json:"asuric_ratio"`       // Percentage shadow / churning resistance
	Demons            []ArishadvargaVector `json:"demons"`             // The 6 Arishadvarga vectors
	GoldenEquilibrium bool                 `json:"golden_equilibrium"` // True if churning tension approaches Φ balance
}

// CalculateTransmutation computes the Samudra Manthan alchemical state with default demons.
func CalculateTransmutation(shadowTension, solarFire float64) TransmutationState {
	return CalculateTransmutationWithDemons(shadowTension, solarFire, nil)
}

// CalculateTransmutationWithDemons computes the Samudra Manthan alchemical state using
// either provided database demon models or canonical defaults.
func CalculateTransmutationWithDemons(shadowTension, solarFire float64, customDemons []db.ArishadvargaDemon) TransmutationState {
	// Clamp values to [0.0, 1.0]
	if shadowTension < 0 {
		shadowTension = 0
	} else if shadowTension > 1 {
		shadowTension = 1
	}
	if solarFire < 0 {
		solarFire = 0
	} else if solarFire > 1 {
		solarFire = 1
	}

	// Dynamic balance:
	// Devic ratio increases with solar fire and decreases with unprocessed shadow.
	devicRaw := (solarFire*0.7 + (1.0-shadowTension)*0.3) * 100.0
	if devicRaw < 10 {
		devicRaw = 10
	} else if devicRaw > 90 {
		devicRaw = 90
	}
	asuricRaw := 100.0 - devicRaw

	dbDemons := customDemons
	if len(dbDemons) != 6 {
		dbDemons = db.DefaultArishadvargaDemons()
	}

	demons := make([]ArishadvargaVector, 6)
	baseDist := 140.0 + shadowTension*70.0

	for i := 0; i < 6; i++ {
		angle := (float64(i) / 6.0) * 2.0 * math.Pi
		cx := baseDist * math.Cos(angle)
		cy := baseDist * math.Sin(angle) + 20.0

		d := dbDemons[i]
		demons[i] = ArishadvargaVector{
			Index:            d.Index,
			Key:              d.Key,
			Name:             d.Name,
			Meaning:          d.ShadowDistortion,
			ShadowDistortion: d.ShadowDistortion,
			AngleRad:         math.Round(angle*1000) / 1000,
			Distance:         math.Round(baseDist*10) / 10,
			Coord:            Point2D{X: math.Round(cx*100) / 100, Y: math.Round(cy*100) / 100},
			TransmutedAs:     d.TransmutedVirtue,
			Teaching:         d.Teaching,
		}
	}

	var state, desc string
	goldenEquilibrium := false

	if solarFire >= 0.75 && shadowTension <= 0.65 {
		state = "pakva"
		desc = "Pakva (पक्व—The Golden Nectar): Under the steady fire of solar awareness, the raw caustic acids have fully ripened into golden sweetness. Karma-Phala is achieved."
		goldenEquilibrium = math.Abs((devicRaw/asuricRaw)-GoldenRatioPhi) < 0.35
	} else if solarFire >= 0.35 {
		state = "manthan"
		desc = "Samudra Manthan (समुद्रमन्थन—The Cosmic Churning): Devic aspiration and Asuric tension are actively rotating Mount Mandara. The friction is transmuting Halahala poison into wisdom."
	} else {
		state = "ama"
		desc = "Āma (आम—Raw Bitter Astringency): Unprocessed shadow and psychic residue predominate. The fruit remains fiercely acidic and resinous, awaiting the fire of Surya."
	}

	return TransmutationState{
		State:             state,
		Description:       desc,
		ShadowTension:     math.Round(shadowTension*100) / 100,
		SolarFire:         math.Round(solarFire*100) / 100,
		DevicRatio:        math.Round(devicRaw*10) / 10,
		AsuricRatio:       math.Round(asuricRaw*10) / 10,
		Demons:            demons,
		GoldenEquilibrium: goldenEquilibrium,
	}
}

// LeafGeometry models one of the 5 sacred mango leaves (Āmra-Pallava)
// crowning the Pūrṇa Kumbha apex.
type LeafGeometry struct {
	Index    int     `json:"index"`
	AngleDeg float64 `json:"angle_deg"`
	Length   float64 `json:"length"`
	SVGPath  string  `json:"svg_path"`
}

// CalculateMangoLeaves generates the 5 sacred mango leaves crowning the apex stem.
func CalculateMangoLeaves(stemOrigin Point2D) []LeafGeometry {
	angles := []float64{-35.0, -18.0, 0.0, 18.0, 35.0}
	lengths := []float64{75.0, 95.0, 110.0, 95.0, 75.0}
	leaves := make([]LeafGeometry, 5)

	for i := 0; i < 5; i++ {
		deg := angles[i]
		rad := (deg - 90.0) * math.Pi / 180.0
		length := lengths[i]

		tipX := stemOrigin.X + length*math.Cos(rad)
		tipY := stemOrigin.Y + length*math.Sin(rad)

		cp1x := stemOrigin.X + (length*0.5)*math.Cos(rad-0.25)
		cp1y := stemOrigin.Y + (length*0.5)*math.Sin(rad-0.25)
		cp2x := stemOrigin.X + (length*0.5)*math.Cos(rad+0.25)
		cp2y := stemOrigin.Y + (length*0.5)*math.Sin(rad+0.25)

		svgPath := fmt.Sprintf("M %.2f %.2f Q %.2f %.2f %.2f %.2f Q %.2f %.2f %.2f %.2f Z",
			stemOrigin.X, stemOrigin.Y,
			cp1x, cp1y, tipX, tipY,
			cp2x, cp2y, stemOrigin.X, stemOrigin.Y)

		leaves[i] = LeafGeometry{
			Index:    i + 1,
			AngleDeg: deg,
			Length:   length,
			SVGPath:  svgPath,
		}
	}
	return leaves
}

// AmraGeometryResponse synthesizes the complete mathematical and alchemical geometry payload.
type AmraGeometryResponse struct {
	Body                 KairiCurveResult           `json:"body"`
	Bija                 KairiCurveResult           `json:"bija"`
	Leaves               []LeafGeometry             `json:"leaves"`
	StemOrigin           Point2D                    `json:"stem_origin"`
	Transmutation        TransmutationState         `json:"transmutation"`
	Philosophy           *db.EsotericPhilosophyDoc  `json:"philosophy,omitempty"`
	Triad                *db.TransmutationTriadDoc  `json:"triad,omitempty"`
	MathematicalFormulas map[string]string          `json:"mathematical_formulas"`
}

// GenerateAmraGeometry produces the unified geometry and alchemical synthesis with default content.
func GenerateAmraGeometry(params KairiParams, shadowTension, solarFire float64) AmraGeometryResponse {
	return GenerateAmraGeometryWithContent(params, shadowTension, solarFire, nil, nil, nil)
}

// GenerateAmraGeometryWithContent produces the unified geometry incorporating database-served esoteric content.
func GenerateAmraGeometryWithContent(
	params KairiParams,
	shadowTension, solarFire float64,
	demons []db.ArishadvargaDemon,
	philosophy *db.EsotericPhilosophyDoc,
	triad *db.TransmutationTriadDoc,
) AmraGeometryResponse {
	body := CalculateKairiCurve(params)

	// Bīja (inner indestructible seed) is calculated with scaled semi-axes and inward focus
	bijaParams := KairiParams{
		Rx:      params.Rx * 0.38,
		Ry:      params.Ry * 0.42,
		Alpha:   params.Alpha * 0.5,
		Beta:    params.Beta * 0.4,
		Gamma:   params.Gamma * 0.3,
		Steps:   60,
		Scale:   0.7,
		OffsetY: params.OffsetY*0.5 + 10.0,
	}
	bija := CalculateKairiCurve(bijaParams)

	stemOrigin := Point2D{X: 0.0, Y: -params.Ry*0.82 + params.OffsetY}
	leaves := CalculateMangoLeaves(stemOrigin)
	transmutation := CalculateTransmutationWithDemons(shadowTension, solarFire, demons)

	formulas := map[string]string{
		"kairi_parametric_x": "x(t) = Scale * (Rx * sin(t) * (1 + α * cos(t)) + γ * ((1 + cos(t))²/4) * (1 - 0.5 * sin(t)))",
		"kairi_parametric_y": "y(t) = Scale * (-Ry * cos(t) + Ry * β * sin²(t/2) * cos(t) + OffsetY)",
		"crest_envelope":     "E(t) = (1 + cos(t))² / 4 [Vanishes at t = ±π, ensuring exact C1 closure]",
		"pratyahara_hook":    "r(θ) = a * exp((ln(Φ) / (π/2)) * θ) [Logarithmic Golden Ratio Spiral]",
		"samudra_manthan":    "Balance = (SolarFire_Deva / ShadowTension_Asura) → GoldenRatio_Φ (1.618)",
		"alchemical_triad":   "Āma (Raw Acidic Poison) ──[Surya Agni]──> Āmra (Karma-Phala Nectar) ──[Essence]──> Amara (Indestructible Bīja)",
	}

	return AmraGeometryResponse{
		Body:                 body,
		Bija:                 bija,
		Leaves:               leaves,
		StemOrigin:           stemOrigin,
		Transmutation:        transmutation,
		Philosophy:           philosophy,
		Triad:                triad,
		MathematicalFormulas: formulas,
	}
}
