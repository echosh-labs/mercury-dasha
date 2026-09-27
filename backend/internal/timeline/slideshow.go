package timeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// HermeticSlide represents a single discrete stanza, scene, and visual asset in an esoteric chronicle.
type HermeticSlide struct {
	Index           int            `json:"index"`                      // 1-indexed order (1, 2, ...)
	Title           string         `json:"title"`                      // Stanza title (e.g. "The Ocean of Light")
	Subtitle        string         `json:"subtitle,omitempty"`        // Epigraph / summary (e.g. "Beyond the Dominion of Night")
	Verses          []string       `json:"verses"`                     // Beautified poem lines
	ImagePath       string         `json:"image_path"`                 // Relative path to asset
	DurationSec     float64        `json:"duration_sec"`               // Display duration in seconds (default: 10.0s)
	Motion          MotionEffect   `json:"motion,omitempty"`          // Ken Burns zoom/pan effect
	Transition      TransitionType `json:"transition,omitempty"`      // Crossfade or Cut
	PlanetaryLord   string         `json:"planetary_lord,omitempty"`   // e.g. "Sun (Surya)", "Mercury (Budha)"
	SacredMetal     string         `json:"sacred_metal,omitempty"`     // e.g. "Quicksilver", "Gold", "Lead"
	HermeticAxiom   string         `json:"hermetic_axiom,omitempty"`   // Principle of Mentalism, Polarity, etc.
	AlchemicalStage string         `json:"alchemical_stage,omitempty"` // Nigredo, Albedo, Citrinitas, Rubedo
	FrequencyHz     float64        `json:"frequency_hz,omitempty"`     // Solfeggio / Root planetary frequency
	ColorHex        string         `json:"color_hex,omitempty"`        // Atmospheric color stamp
}

// HermeticSlideshow is the master declarative document for a repeatable spiritual/character slideshow.
type HermeticSlideshow struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Subtitle         string          `json:"subtitle,omitempty"`
	Author           string          `json:"author,omitempty"`
	Tradition        string          `json:"tradition,omitempty"`
	Description      string          `json:"description,omitempty"`
	Orientation      Orientation     `json:"orientation"`
	TotalDurationSec float64         `json:"total_duration_sec"`
	Slides           []HermeticSlide `json:"slides"`
	VoiceAudioPath   string          `json:"voice_audio_path,omitempty"`
	AmbientMusicPath string          `json:"ambient_music_path,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	Metadata         map[string]any  `json:"metadata,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// SlideshowOptions specifies runtime parameters when generating a timeline manifest.
type SlideshowOptions struct {
	Orientation             Orientation // 16:9, 9:16, 1:1
	DefaultSlideDurationSec float64     // Fallback duration if slide.DurationSec is 0 (default: 10.0s)
	VoiceAudioPath          string      // Optional spoken narrative
	AmbientMusicPath        string      // Optional background drone
	IncludeOverlays         bool        // Attach ChronoBadge, LowerThird, Waveform
	EnableDucking           bool        // Duck ambient audio when voice is active
	AmbientVolume           float64     // Default 0.25
}

// EnsureDefaults applies sensible defaults to the slideshow and its slides.
func (s *HermeticSlideshow) EnsureDefaults() {
	if s.Orientation == "" {
		s.Orientation = OrientationLandscape16x9
	}
	if s.Tags == nil {
		s.Tags = make([]string, 0)
	}
	if s.Metadata == nil {
		s.Metadata = make(map[string]any)
	}

	motionCycle := []MotionEffect{
		MotionKenBurnsIn,
		MotionPanLeft,
		MotionKenBurnsOut,
		MotionPanRight,
	}

	var totalDur float64
	for i := range s.Slides {
		slide := &s.Slides[i]
		if slide.Index <= 0 {
			slide.Index = i + 1
		}
		if slide.DurationSec <= 0 {
			slide.DurationSec = 10.0
		}
		totalDur += slide.DurationSec

		if slide.Motion == "" {
			slide.Motion = motionCycle[i%len(motionCycle)]
		}
		if slide.Transition == "" {
			slide.Transition = TransitionCrossfade
		}
	}
	s.TotalDurationSec = totalDur

	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
}

// GenerateTimelineManifest converts the HermeticSlideshow into a production-ready TimelineManifest.
func (s *HermeticSlideshow) GenerateTimelineManifest(opts SlideshowOptions) (*TimelineManifest, error) {
	if s == nil {
		return nil, errors.New("slideshow cannot be nil")
	}
	s.EnsureDefaults()

	duration := s.TotalDurationSec
	if duration <= 0 {
		duration = float64(len(s.Slides)) * 10.0
		if duration <= 0 {
			duration = 60.0
		}
	}

	manifestID := fmt.Sprintf("manifest-%s-%d", s.ID, time.Now().Unix())
	manifest := NewTimelineManifest(manifestID, s.Title)
	manifest.Description = s.Description
	manifest.Tags = append(manifest.Tags, s.Tags...)
	manifest.Tags = append(manifest.Tags, "hermetic_slideshow", "audiovisual_render", strings.ToLower(s.Tradition))

	// Canvas orientation
	if opts.Orientation != "" {
		manifest.Canvas.Orientation = opts.Orientation
	} else if s.Orientation != "" {
		manifest.Canvas.Orientation = s.Orientation
	}
	manifest.Canvas.EnsureDefaults()

	// 1. ChronoBinding Metaphysics Stamping
	firstAxiom := "Principle of Mentalism"
	firstMetal := "Quicksilver"
	firstLord := "Surya (Sun)"
	firstStage := "Albedo"
	if len(s.Slides) > 0 {
		if s.Slides[0].HermeticAxiom != "" {
			firstAxiom = s.Slides[0].HermeticAxiom
		}
		if s.Slides[0].SacredMetal != "" {
			firstMetal = s.Slides[0].SacredMetal
		}
		if s.Slides[0].PlanetaryLord != "" {
			firstLord = s.Slides[0].PlanetaryLord
		}
		if s.Slides[0].AlchemicalStage != "" {
			firstStage = s.Slides[0].AlchemicalStage
		}
	}

	manifest.ChronoMeta = &ChronoBinding{
		VimshottariLord: firstLord,
		PlanetaryHora:   firstLord,
		SacredMetal:     firstMetal,
		HermeticAxiom:   firstAxiom,
		MagnumOpusStage: firstStage,
	}

	// 2. Audio Composition
	voicePath := s.VoiceAudioPath
	if opts.VoiceAudioPath != "" {
		voicePath = opts.VoiceAudioPath
	}

	var voiceTrackID string
	if voicePath != "" {
		voiceTrackID = fmt.Sprintf("voice-%s", s.ID)
		manifest.AddAudioTrack(AudioTrack{
			ID:         voiceTrackID,
			Role:       AudioRoleVoice,
			SourcePath: voicePath,
			StartSec:   0,
			EndSec:     duration,
			Volume:     1.0,
		})
	}

	ambientPath := s.AmbientMusicPath
	if opts.AmbientMusicPath != "" {
		ambientPath = opts.AmbientMusicPath
	}
	if ambientPath == "" {
		ambientPath = "drone_528hz_solfeggio.wav" // Standard 528 Hz transformation frequency bed
	}

	ambientVol := opts.AmbientVolume
	if ambientVol <= 0 {
		ambientVol = 0.28
	}

	ambientTrack := AudioTrack{
		ID:         fmt.Sprintf("ambient-%s", s.ID),
		Role:       AudioRoleAmbientBed,
		SourcePath: ambientPath,
		StartSec:   0,
		EndSec:     duration,
		Volume:     ambientVol,
		FadeInSec:  2.0,
		FadeOutSec: 3.0,
		Loop:       true,
	}

	if voiceTrackID != "" || opts.EnableDucking {
		ambientTrack.Ducking = &AudioDucking{
			Enabled:       true,
			AttenuateTo:   0.12,
			AttackMs:      150,
			ReleaseMs:     700,
			SidechainFrom: voiceTrackID,
		}
	}
	manifest.AddAudioTrack(ambientTrack)

	// 3. Sequential Visual Scenes from Slides
	var currentStart float64
	for i, slide := range s.Slides {
		slideDur := slide.DurationSec
		if slideDur <= 0 {
			slideDur = 10.0
		}

		subtitle := slide.Subtitle
		if subtitle == "" && len(slide.Verses) > 0 {
			subtitle = slide.Verses[0]
		}

		scene := VisualScene{
			ID:            fmt.Sprintf("scene-%02d-%s", slide.Index, sanitizeSlug(slide.Title)),
			StartSec:      currentStart,
			DurationSec:   slideDur,
			SourceType:    "image",
			SourcePath:    slide.ImagePath,
			Motion:        slide.Motion,
			MotionScale:   1.08,
			TransitionIn:  slide.Transition,
			TransitionSec: 1.2,
			Caption:       slide.Title,
			Subtitle:      subtitle,
		}
		if i == 0 {
			scene.TransitionIn = TransitionFadeToBlack
			scene.TransitionSec = 1.5
		}

		manifest.AddScene(scene)
		currentStart += slideDur
	}

	// 4. Overlays
	if opts.IncludeOverlays {
		// Chrono Badge in top right
		manifest.AddOverlay(GlobalOverlay{
			ID:       "overlay-chrono-badge",
			Type:     OverlayChronoBadge,
			Position: "top_right",
			StartSec: 0,
			EndSec:   duration,
			Config: map[string]interface{}{
				"title":        s.Title,
				"sacred_metal": firstMetal,
				"axiom":        firstAxiom,
				"stage":        firstStage,
			},
		})

		// Lower third presenter / chronicle attribution
		manifest.AddOverlay(GlobalOverlay{
			ID:       "overlay-lower-third",
			Type:     OverlayLowerThird,
			Position: "bottom",
			StartSec: 2.0,
			EndSec:   duration - 2.0,
			Config: map[string]interface{}{
				"title":     s.Title,
				"subtitle":  s.Subtitle,
				"tradition": s.Tradition,
			},
		})

		// Waveform visualizer at bottom center
		manifest.AddOverlay(GlobalOverlay{
			ID:       "overlay-waveform",
			Type:     OverlayWaveform,
			Position: "bottom_center",
			StartSec: 0,
			EndSec:   duration,
			Config: map[string]interface{}{
				"color_hex": "#d4af37", // Imperial Gold
				"opacity":   0.85,
			},
		})
	}

	manifest.DurationSec = currentStart
	manifest.UpdatedAt = time.Now().UTC()

	return manifest, nil
}

// ToIndexEntries converts the slideshow and its slide images into standard Sovereign Storehouse IndexEntries.
func (s *HermeticSlideshow) ToIndexEntries(baseStorePath string) []db.IndexEntry {
	entries := make([]db.IndexEntry, 0, len(s.Slides)+1)
	now := time.Now().UTC()

	// 1. Slideshow Manifest Entry
	manifestRelPath := fmt.Sprintf("slideshows/%s/manifest.json", s.ID)
	entries = append(entries, db.IndexEntry{
		ID:        manifestRelPath,
		Category:  "code",
		Path:      manifestRelPath,
		FullPath:  filepath.Join(baseStorePath, filepath.FromSlash(manifestRelPath)),
		FileName:  "manifest.json",
		Extension: ".json",
		SizeBytes: 4096,
		ModTime:   now,
		IndexedAt: now,
		Tags: []string{
			"slideshow",
			"hermetic_story",
			"timeline_manifest",
			"tradition:" + strings.ToLower(s.Tradition),
		},
		Snippet: s.Description,
		Metadata: map[string]any{
			"slideshow_id":       s.ID,
			"title":              s.Title,
			"author":             s.Author,
			"total_duration_sec": s.TotalDurationSec,
			"slide_count":        len(s.Slides),
		},
	})

	// 2. Individual Slide Image Entries
	for _, slide := range s.Slides {
		ext := filepath.Ext(slide.ImagePath)
		fileName := filepath.Base(slide.ImagePath)
		relPath := strings.TrimPrefix(slide.ImagePath, "/")

		slideTags := []string{
			"slideshow_slide",
			"image",
			"hermetic_art",
			fmt.Sprintf("slideshow:%s", s.ID),
			fmt.Sprintf("stanza:%d", slide.Index),
		}
		if slide.SacredMetal != "" {
			slideTags = append(slideTags, "metal:"+strings.ToLower(slide.SacredMetal))
		}
		if slide.PlanetaryLord != "" {
			slideTags = append(slideTags, "planet:"+strings.ToLower(slide.PlanetaryLord))
		}
		if slide.AlchemicalStage != "" {
			slideTags = append(slideTags, "stage:"+strings.ToLower(slide.AlchemicalStage))
		}

		entries = append(entries, db.IndexEntry{
			ID:        relPath,
			Category:  "video", // Visual scene asset for video generation
			Path:      relPath,
			FullPath:  filepath.Join(baseStorePath, filepath.FromSlash(relPath)),
			FileName:  fileName,
			Extension: ext,
			SizeBytes: 1024 * 1024, // Nominal 1MB
			ModTime:   now,
			IndexedAt: now,
			Tags:      slideTags,
			Snippet:   strings.Join(slide.Verses, "\n"),
			Metadata: map[string]any{
				"slideshow_id":     s.ID,
				"stanza_index":     slide.Index,
				"title":            slide.Title,
				"subtitle":         slide.Subtitle,
				"sacred_metal":     slide.SacredMetal,
				"hermetic_axiom":   slide.HermeticAxiom,
				"planetary_lord":   slide.PlanetaryLord,
				"alchemical_stage": slide.AlchemicalStage,
				"frequency_hz":     slide.FrequencyHz,
			},
		})
	}

	return entries
}

// SaveHermeticSlideshow persists a HermeticSlideshow into BoltDB esoteric_content bucket.
func SaveHermeticSlideshow(store db.StorageEngine, s *HermeticSlideshow) error {
	if store == nil {
		return errors.New("store cannot be nil")
	}
	if s == nil || s.ID == "" {
		return errors.New("slideshow or slideshow ID cannot be empty")
	}
	s.EnsureDefaults()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal hermetic slideshow: %w", err)
	}

	key := "slideshow:" + s.ID
	return store.SaveEsotericContent(key, data)
}

// GetHermeticSlideshow retrieves a HermeticSlideshow from BoltDB.
func GetHermeticSlideshow(store db.StorageEngine, id string) (*HermeticSlideshow, error) {
	if store == nil {
		return nil, errors.New("store cannot be nil")
	}
	cleanID := strings.TrimPrefix(id, "slideshow:")
	key := "slideshow:" + cleanID

	raw, err := store.GetEsotericContent(key)
	if err != nil {
		return nil, err
	}

	var s HermeticSlideshow
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal slideshow JSON: %w", err)
	}
	s.EnsureDefaults()
	return &s, nil
}

// ListHermeticSlideshows enumerates all stored hermetic slideshows.
func ListHermeticSlideshows(store db.StorageEngine) ([]*HermeticSlideshow, error) {
	if store == nil {
		return nil, errors.New("store cannot be nil")
	}

	keys, err := store.ListEsotericContent("slideshow:")
	if err != nil {
		return nil, err
	}

	results := make([]*HermeticSlideshow, 0, len(keys))
	for _, key := range keys {
		raw, err := store.GetEsotericContent(key)
		if err != nil {
			continue
		}
		var s HermeticSlideshow
		if err := json.Unmarshal(raw, &s); err == nil {
			s.EnsureDefaults()
			results = append(results, &s)
		}
	}
	return results, nil
}

// IndexSlideshowAssets writes index entries for the slideshow into the sovereign storehouse index.
func IndexSlideshowAssets(store db.StorageEngine, s *HermeticSlideshow, baseStorePath string) error {
	if store == nil || s == nil {
		return errors.New("store and slideshow cannot be nil")
	}
	entries := s.ToIndexEntries(baseStorePath)
	return store.BatchPutIndexEntries(entries)
}

// NewTempleOfIlluminationSlideshow instantiates the canonical 8-stanza Temple of Illumination chronicle.
func NewTempleOfIlluminationSlideshow() *HermeticSlideshow {
	slideshow := &HermeticSlideshow{
		ID:          "temple-of-illumination",
		Title:       "The Temple of Illumination",
		Subtitle:    "The Ocean of Light & The Mount of the Wise",
		Author:      "Sovereign Genesis / Esoteric Hermetic Canon",
		Tradition:   "Hermetic-Rosicrucian Esotericism",
		Description: "A sacred 8-scene mystical chronicle depicting the soul's ascent across the boundless Ocean of Light to the Island of Blessing, the living fountain atop the Mount of the Wise, and the innermost sanctuary of the House of the Sun.",
		Orientation: OrientationLandscape16x9,
		Tags: []string{
			"temple_of_illumination",
			"house_of_the_sun",
			"mount_of_the_wise",
			"ocean_of_light",
			"hermetic_chronicle",
			"sacred_poetry",
			"eternal_life",
		},
		Metadata: map[string]any{
			"stanzas_count": 8,
			"master_axiom":  "Principle of Mentalism: The All is Mind; the Universe is Mental.",
			"primary_metal": "Solar Electrum / Quicksilver Bridge",
		},
		Slides: []HermeticSlide{
			{
				Index:    1,
				Title:    "The Ocean of Light",
				Subtitle: "Beyond the Dominion of Night",
				Verses: []string{
					"There’s an ocean that men cannot fathom nor measure;",
					"It lies just beyond the Dominion of Night;",
					"’Tis the ocean of splendor, of infinite pleasure,",
					"Of fathomless beauty—the ocean of Light.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/01_ocean_of_light.jpg",
				DurationSec:     12.0,
				Motion:          MotionKenBurnsIn,
				Transition:      TransitionFadeToBlack,
				PlanetaryLord:   "Mercury (Budha)",
				SacredMetal:     "Quicksilver",
				HermeticAxiom:   "Principle of Polarity",
				AlchemicalStage: "Nigredo to Albedo",
				FrequencyHz:     528.0,
				ColorHex:        "#e6c35c",
			},
			{
				Index:    2,
				Title:    "The Island of Blessing",
				Subtitle: "The Gem of the Sea & Home of the Spirit",
				Verses: []string{
					"In the midst of this radiant ocean of glory",
					"Rests the Island of Blessing, the gem of the sea;",
					"The home of the Spirit, so famous in story,",
					"Where angels are servants, and men are the free.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/02_island_of_blessing.jpg",
				DurationSec:     12.0,
				Motion:          MotionPanLeft,
				Transition:      TransitionCrossfade,
				PlanetaryLord:   "Moon (Chandra)",
				SacredMetal:     "Silver",
				HermeticAxiom:   "Principle of Correspondence",
				AlchemicalStage: "Albedo",
				FrequencyHz:     432.0,
				ColorHex:        "#5cdb95",
			},
			{
				Index:    3,
				Title:    "The Mount of the Wise",
				Subtitle: "The Flower-Crowned Summit & Fountain of Life",
				Verses: []string{
					"In the midst of the Isle is a flower-crowned mountain;",
					"The sanctified call it the Mount of the Wise;",
					"From its summit pours forth a life-giving fountain",
					"That waters the lands of the earth and the skies.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/03_mount_of_wise.jpg",
				DurationSec:     12.0,
				Motion:          MotionKenBurnsOut,
				Transition:      TransitionCrossfade,
				PlanetaryLord:   "Jupiter (Brihaspati)",
				SacredMetal:     "Tin",
				HermeticAxiom:   "Principle of Vibration",
				AlchemicalStage: "Citrinitas",
				FrequencyHz:     639.0,
				ColorHex:        "#38ef7d",
			},
			{
				Index:    4,
				Title:    "The House of the Sun",
				Subtitle: "Ten Thousand Angels & Souls Victorious",
				Verses: []string{
					"On the top of the mountain a Temple, all glorious,",
					"Stands out in the light of the Illumined One.",
					"Ten thousand bright angels and souls all victorious,",
					"Surround it, and fill it—this House of the Sun.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/04_house_of_sun.jpg",
				DurationSec:     12.0,
				Motion:          MotionPanRight,
				Transition:      TransitionCrossfade,
				PlanetaryLord:   "Sun (Surya)",
				SacredMetal:     "Gold",
				HermeticAxiom:   "Principle of Gender",
				AlchemicalStage: "Rubedo",
				FrequencyHz:     126.22,
				ColorHex:        "#ffd700",
			},
			{
				Index:    5,
				Title:    "The Temple of Illumination",
				Subtitle: "The Grand Council of Heaven and Earth",
				Verses: []string{
					"And this is the Temple of Illumination",
					"Where the courtiers of heaven and earth daily meet;",
					"Where souls, cleansed from sin, elect from each nation,",
					"Hold council with Jesus, and sit at his feet.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/05_temple_of_illumination.jpg",
				DurationSec:     12.0,
				Motion:          MotionKenBurnsIn,
				Transition:      TransitionCrossfade,
				PlanetaryLord:   "Sun (Surya) / Christic Logos",
				SacredMetal:     "Gold",
				HermeticAxiom:   "Principle of Mentalism",
				AlchemicalStage: "Rubedo",
				FrequencyHz:     741.0,
				ColorHex:        "#f7971e",
			},
			{
				Index:    6,
				Title:    "The Valley of Silence and Prayer",
				Subtitle: "The Contemplative Way Unto the Mountain",
				Verses: []string{
					"The way to this Island and unto this Mountain",
					"Lies through the deep valley of Silence and Prayer;",
					"But whoever will may drink from the Fountain,",
					"And realize all that it is to be there.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/06_valley_of_silence.jpg",
				DurationSec:     12.0,
				Motion:          MotionPanLeft,
				Transition:      TransitionCrossfade,
				PlanetaryLord:   "Saturn (Shani)",
				SacredMetal:     "Lead",
				HermeticAxiom:   "Principle of Rhythm",
				AlchemicalStage: "Nigredo",
				FrequencyHz:     396.0,
				ColorHex:        "#8e2de2",
			},
			{
				Index:    7,
				Title:    "The Thrice-Blessed Day",
				Subtitle: "Universal Welcome & The Radiant Dawn",
				Verses: []string{
					"Come up to this Temple of Illumination;",
					"Come, bathe in the sunlight of a thrice-blessed day.",
					"There is room for the millions of every nation,",
					"And Christ is the Truth and the Life and the Way.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/07_thrice_blessed_day.jpg",
				DurationSec:     12.0,
				Motion:          MotionKenBurnsOut,
				Transition:      TransitionCrossfade,
				PlanetaryLord:   "Venus (Shukra)",
				SacredMetal:     "Copper",
				HermeticAxiom:   "Principle of Cause and Effect",
				AlchemicalStage: "Citrinitas",
				FrequencyHz:     852.0,
				ColorHex:        "#ff6a00",
			},
			{
				Index:    8,
				Title:    "The Innermost Circle",
				Subtitle: "The Angel of Mercy & Eternal Benediction",
				Verses: []string{
					"In the Innermost Circle there’s joy and there’s gladness;",
					"There’s peace and there’s freedom from sin and from strife.",
					"The Angel of Mercy will free you from sadness,",
					"And Christ’s Benediction is Eternal Life.",
				},
				ImagePath:       "assets/slideshows/temple_of_illumination/08_innermost_circle.jpg",
				DurationSec:     14.0,
				Motion:          MotionKenBurnsIn,
				Transition:      TransitionCrossfade,
				PlanetaryLord:   "Ketu / Divine Grace",
				SacredMetal:     "Philosopher's Stone / Amara",
				HermeticAxiom:   "Principle of Mental Transmutation",
				AlchemicalStage: "Magnum Opus Complete",
				FrequencyHz:     963.0,
				ColorHex:        "#ffffff",
			},
		},
	}
	slideshow.EnsureDefaults()
	return slideshow
}

// SeedTempleOfIlluminationSlideshow seeds the canonical Temple of Illumination slideshow into Hermetic storage and storehouse index.
func SeedTempleOfIlluminationSlideshow(store db.StorageEngine, baseStorePath string) error {
	if store == nil {
		return errors.New("store cannot be nil")
	}
	s := NewTempleOfIlluminationSlideshow()
	if _, err := GetHermeticSlideshow(store, s.ID); errors.Is(err, db.ErrNotFound) {
		if err := SaveHermeticSlideshow(store, s); err != nil {
			return fmt.Errorf("failed to seed hermetic slideshow: %w", err)
		}
		if err := IndexSlideshowAssets(store, s, baseStorePath); err != nil {
			return fmt.Errorf("failed to index slideshow assets: %w", err)
		}
	}
	return nil
}

func sanitizeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "’", "")
	s = strings.ReplaceAll(s, "—", "-")
	s = strings.ReplaceAll(s, "--", "-")
	return strings.Trim(s, "-")
}
