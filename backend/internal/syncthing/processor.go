package syncthing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
)

// MediaProcessor parses incoming mobile media, enriches it with astrological
// Vimshottari Dasha periods and video metadata, and indexes it into BoltDB.
type MediaProcessor struct {
	store        db.StorageEngine
	timeline     []dasha.DashaPeriod
	mediaRoot    string
	thumbnailDir string

	feedMu sync.RWMutex
	feed   []RecentMediaItem
}

// NewMediaProcessor instantiates a media pipeline worker.
func NewMediaProcessor(store db.StorageEngine, timeline []dasha.DashaPeriod, mediaRoot, thumbnailDir string) *MediaProcessor {
	if mediaRoot == "" {
		mediaRoot = "/home/justin/media"
	}
	if thumbnailDir == "" {
		thumbnailDir = filepath.Join(mediaRoot, ".thumbnails")
	}
	_ = os.MkdirAll(thumbnailDir, 0755)

	return &MediaProcessor{
		store:        store,
		timeline:     timeline,
		mediaRoot:    mediaRoot,
		thumbnailDir: thumbnailDir,
		feed:         make([]RecentMediaItem, 0, 100),
	}
}

// UpdateTimeline updates the astrological Dasha timeline snapshot.
func (p *MediaProcessor) UpdateTimeline(timeline []dasha.DashaPeriod) {
	p.timeline = timeline
}

// ProcessMediaFile ingests, enriches, and catalogs a photo or video.
func (p *MediaProcessor) ProcessMediaFile(ctx context.Context, fullPath string, folderID string) (*RecentMediaItem, error) {
	fi, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}
	if fi.IsDir() {
		return nil, nil // skip directories
	}

	fileName := filepath.Base(fullPath)
	ext := strings.ToLower(filepath.Ext(fileName))

	isPhoto := indexer.IsPhotoExtension(ext)
	isVideo := indexer.IsVideoExtension(ext)

	if !isPhoto && !isVideo {
		return nil, nil // skip non-media files
	}

	// Calculate canonical relative path
	var relPath string
	if rel, err := filepath.Rel(p.mediaRoot, fullPath); err == nil && !strings.HasPrefix(rel, "..") {
		relPath = rel
	} else {
		matched := false
		if entries, err := os.ReadDir(p.mediaRoot); err == nil {
			for _, entry := range entries {
				entryPath := filepath.Join(p.mediaRoot, entry.Name())
				if realTarget, err := filepath.EvalSymlinks(entryPath); err == nil {
					if subRel, err := filepath.Rel(realTarget, fullPath); err == nil && !strings.HasPrefix(subRel, "..") {
						relPath = filepath.Join(entry.Name(), subRel)
						matched = true
						break
					}
				}
			}
		}
		if !matched {
			relPath = fileName
		}
	}
	relPath = filepath.ToSlash(relPath)

	category := "photos"
	if isVideo {
		category = "video"
	}

	// 1. Extract Media Timestamp
	mediaDate, hasExplicitDate := indexer.ExtractPhotoDate(relPath, fileName, fi.ModTime())

	// 2. Video / Photo Metadata Extraction
	var durationSec float64
	var resolution string
	var deviceModel string
	var location string
	var thumbFile string

	if isVideo {
		vMeta := p.extractVideoMetadata(ctx, fullPath)
		if vMeta != nil {
			durationSec = vMeta.DurationSec
			resolution = vMeta.Resolution
			deviceModel = vMeta.DeviceModel
			location = vMeta.Location
			if !hasExplicitDate && !vMeta.CreationTime.IsZero() {
				mediaDate = vMeta.CreationTime
			}
		}

		// Generate Video Thumbnail via ffmpeg
		thumbFile = p.generateVideoThumbnail(ctx, fullPath, fileName)
	} else if isPhoto {
		// Photos act as their own thumbnail or use full file path
		thumbFile = fileName
	}

	// 3. Resolve Astrological Vimshottari Dasha Snapshot
	var mahaName, antarName, metalName, metalLatin, horaName, hermeticAxiom string
	if len(p.timeline) > 0 {
		snap := dasha.ResolveActiveSnapshot(p.timeline, mediaDate)
		mahaName = snap.Mahadasha.PlanetName
		antarName = snap.Antardasha.PlanetName
		mName, mLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)
		metalName = mName
		metalLatin = mLatin
		horaName = antarName // Antardasha sub-ruler resonance
		hermeticAxiom = fmt.Sprintf("Governed under %s Mahadasha (%s)", mahaName, metalName)
	}

	// 4. Construct Tags & Metadata Map
	tags := []string{
		category,
		strings.TrimPrefix(ext, "."),
		"phone",
		"media_vault",
		fmt.Sprintf("year:%d", mediaDate.Year()),
	}
	if mahaName != "" {
		tags = append(tags, strings.ToLower("mahadasha:"+mahaName))
	}
	if metalName != "" {
		tags = append(tags, strings.ToLower("metal:"+metalName))
	}
	if isVideo {
		tags = append(tags, "video_stream")
	}

	metadata := map[string]any{
		"album":            filepath.Base(filepath.Dir(fullPath)),
		"photo_date":       mediaDate.Format("2006-01-02"),
		"year":             mediaDate.Year(),
		"has_explicit":     hasExplicitDate,
		"dasha_mahadasha":  mahaName,
		"dasha_antardasha": antarName,
		"sacred_metal":     metalName + " (" + metalLatin + ")",
		"folder_id":        folderID,
	}
	if durationSec > 0 {
		metadata["duration_sec"] = durationSec
	}
	if resolution != "" {
		metadata["resolution"] = resolution
	}
	if deviceModel != "" {
		metadata["device_model"] = deviceModel
	}
	if location != "" {
		metadata["location"] = location
	}

	// 5. Store in BoltDB Catalog Index
	entryID := "media/" + relPath
	entry := db.IndexEntry{
		ID:        entryID,
		Category:  category,
		Path:      relPath,
		FullPath:  fullPath,
		FileName:  fileName,
		Extension: ext,
		SizeBytes: fi.Size(),
		ModTime:   fi.ModTime(),
		IndexedAt: time.Now().UTC(),
		Tags:      tags,
		Metadata:  metadata,
	}

	if p.store != nil {
		if err := p.store.BatchPutIndexEntries([]db.IndexEntry{entry}); err != nil {
			log.Printf("[Syncthing] Error indexing media entry %s: %v", entryID, err)
		}
	}

	// 6. Build High-Level RecentMediaItem & Update In-Memory Feed
	thumbURL := ""
	if isVideo && thumbFile != "" {
		thumbURL = "/api/v1/media/thumbnail?name=" + thumbFile
	} else if isPhoto {
		thumbURL = "/api/v1/media/stream?path=" + relPath
	}
	streamURL := "/api/v1/media/stream?path=" + relPath

	item := RecentMediaItem{
		ID:              entryID,
		FileName:        fileName,
		Category:        category,
		Path:            relPath,
		FullPath:        fullPath,
		SizeBytes:       fi.Size(),
		ModTime:         fi.ModTime(),
		DurationSec:     durationSec,
		Resolution:      resolution,
		ThumbnailURL:    thumbURL,
		StreamURL:       streamURL,
		Tags:            tags,
		DashaMahadasha:  mahaName,
		DashaAntardasha: antarName,
		SacredMetal:     metalName + " (" + metalLatin + ")",
		Hora:            horaName,
		HermeticAxiom:   hermeticAxiom,
		DeviceModel:     deviceModel,
		Location:        location,
		SyncedAt:        time.Now().UTC(),
		Metadata:        metadata,
	}

	p.addToFeed(item)
	return &item, nil
}

func (p *MediaProcessor) addToFeed(item RecentMediaItem) {
	p.feedMu.Lock()
	defer p.feedMu.Unlock()

	// Prepend
	p.feed = append([]RecentMediaItem{item}, p.feed...)
	if len(p.feed) > 100 {
		p.feed = p.feed[:100]
	}
}

// GetRecentFeed returns up to limit recent items from memory.
func (p *MediaProcessor) GetRecentFeed(limit int) []RecentMediaItem {
	p.feedMu.RLock()
	defer p.feedMu.RUnlock()

	if limit <= 0 || limit > len(p.feed) {
		limit = len(p.feed)
	}
	result := make([]RecentMediaItem, limit)
	copy(result, p.feed[:limit])
	return result
}

type videoProbeMeta struct {
	DurationSec  float64
	Resolution   string
	DeviceModel  string
	Location     string
	CreationTime time.Time
}

func (p *MediaProcessor) extractVideoMetadata(ctx context.Context, fullPath string) *videoProbeMeta {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		fullPath,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var data struct {
		Format struct {
			Duration string `json:"duration"`
			Tags     struct {
				CreationTime string `json:"creation_time"`
				Location     string `json:"location"`
				Model        string `json:"com.android.model"`
			} `json:"tags"`
		} `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(out, &data); err != nil {
		return nil
	}

	meta := &videoProbeMeta{
		DeviceModel: data.Format.Tags.Model,
		Location:    data.Format.Tags.Location,
	}

	if d, err := strconv.ParseFloat(data.Format.Duration, 64); err == nil {
		meta.DurationSec = d
	}

	if t, err := time.Parse(time.RFC3339, data.Format.Tags.CreationTime); err == nil {
		meta.CreationTime = t
	}

	for _, s := range data.Streams {
		if s.CodecType == "video" && s.Width > 0 && s.Height > 0 {
			meta.Resolution = fmt.Sprintf("%dx%d", s.Width, s.Height)
			break
		}
	}

	return meta
}

func (p *MediaProcessor) generateVideoThumbnail(ctx context.Context, fullPath, fileName string) string {
	hasher := sha256.New()
	hasher.Write([]byte(fullPath))
	hashStr := hex.EncodeToString(hasher.Sum(nil))[:16]
	thumbFileName := fmt.Sprintf("thumb_%s.jpg", hashStr)
	thumbPath := filepath.Join(p.thumbnailDir, thumbFileName)

	if _, err := os.Stat(thumbPath); err == nil {
		return thumbFileName // already exists
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-ss", "00:00:01",
		"-i", fullPath,
		"-vframes", "1",
		"-q:v", "2",
		thumbPath,
	)

	if err := cmd.Run(); err != nil {
		log.Printf("[Syncthing] Thumbnail generation failed for %s: %v", fileName, err)
		return ""
	}

	return thumbFileName
}
