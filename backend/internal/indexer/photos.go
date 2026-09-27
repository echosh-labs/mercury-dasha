package indexer

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var (
	rePXLDate = regexp.MustCompile(`(?i)\bPXL_(\d{4})(\d{2})(\d{2})_`)
	reIMGDate = regexp.MustCompile(`(?i)\bIMG_(\d{4})(\d{2})(\d{2})_`)
)

// IsPhotoExtension returns true if the extension matches a common image format.
func IsPhotoExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".heic":
		return true
	default:
		return false
	}
}

// ExtractPhotoDate attempts to determine the date of a photo from its filename,
// parent album folder, or modification timestamp.
func ExtractPhotoDate(relPath, fileName string, modTime time.Time) (time.Time, bool) {
	// 1. Check Google Pixel pattern: PXL_20260917_...
	if m := rePXLDate.FindStringSubmatch(fileName); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		if y >= 1970 && y <= 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 {
			return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, time.UTC), true
		}
	}

	// 2. Check IMG pattern: IMG_20200512_...
	if m := reIMGDate.FindStringSubmatch(fileName); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		if y >= 1970 && y <= 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 {
			return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, time.UTC), true
		}
	}

	// 3. Check ISO pattern in filename: 2013-05-21 or 2013_05_21
	if m := reDateDash.FindStringSubmatch(fileName); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		if y >= 1970 && y <= 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 {
			return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, time.UTC), true
		}
	}

	// 4. Check folder name for Year (e.g. "Pictures/Sarah's Birth 2009" or "Pictures/080401")
	dir := filepath.Dir(relPath)
	if m := reYearOnly.FindStringSubmatch(dir); len(m) == 2 {
		y, _ := strconv.Atoi(m[1])
		if y >= 1970 && y <= 2088 {
			return time.Date(y, time.January, 1, 12, 0, 0, 0, time.UTC), true
		}
	}

	// Fallback to ModTime
	return modTime, false
}

// EnhancePhotoEntry annotates a photo IndexEntry with album, date, and Dasha period.
func EnhancePhotoEntry(entry *db.IndexEntry, timeline []dasha.DashaPeriod) {
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]any)
	}

	photoDate, hasExplicit := ExtractPhotoDate(entry.Path, entry.FileName, entry.ModTime)
	entry.Metadata["photo_date"] = photoDate.Format("2006-01-02")
	entry.Metadata["year"] = photoDate.Year()
	entry.Metadata["has_explicit_date"] = hasExplicit

	// Determine Album Name
	dir := filepath.Dir(entry.Path)
	album := filepath.Base(dir)
	if album == "." || album == "Pictures" || album == "images" {
		album = "General Archive"
	}
	entry.Metadata["album"] = album

	entry.Tags = append(entry.Tags,
		"photo",
		"image",
		"year:"+strconv.Itoa(photoDate.Year()),
		"album:"+strings.ToLower(album),
	)

	// Astrological Dasha groundings
	if len(timeline) > 0 {
		snap := dasha.ResolveActiveSnapshot(timeline, photoDate)
		mahaGraha := snap.Mahadasha.Planet
		metalName, metalLatin := PlanetSacredMetal(mahaGraha)

		entry.Metadata["dasha_mahadasha"] = snap.Mahadasha.PlanetName
		entry.Metadata["dasha_antardasha"] = snap.Antardasha.PlanetName
		entry.Metadata["sacred_metal"] = metalName + " (" + metalLatin + ")"

		entry.Tags = append(entry.Tags,
			"mahadasha:"+strings.ToLower(snap.Mahadasha.PlanetName),
			"metal:"+strings.ToLower(metalName),
		)
	}

	entry.Tags = dedupeStrings(entry.Tags)
}
