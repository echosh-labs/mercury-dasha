package indexer

import (
	"path/filepath"
	"strings"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// ParseBookMetadata extracts intelligent taxonomy, author, and title tags from book paths.
func ParseBookMetadata(relPath string, fileName string, sizeBytes int64) (tags []string, meta map[string]any) {
	tags = []string{"book"}
	meta = make(map[string]any)

	// Clean base filename without extension
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)

	// Extract author and title if filename contains " - "
	if strings.Contains(base, " - ") {
		parts := strings.SplitN(base, " - ", 2)
		author := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])
		if len(author) > 1 && len(title) > 1 {
			meta["author"] = author
			meta["title"] = title
			tags = append(tags, strings.ToLower(author))
		}
	} else {
		meta["title"] = base
	}

	// Extract collection & subjects from path hierarchy
	// e.g. "books/Guild Archives/Rasa Shastra.pdf" -> collection: "Guild Archives"
	dirs := strings.Split(filepath.Dir(relPath), string(filepath.Separator))
	for _, d := range dirs {
		dClean := strings.TrimSpace(d)
		if dClean == "" || strings.EqualFold(dClean, "books") || dClean == "." {
			continue
		}
		tags = append(tags, strings.ToLower(dClean))
		if meta["collection"] == nil {
			meta["collection"] = dClean
		}
	}

	// Format tags
	switch strings.ToLower(ext) {
	case ".pdf":
		tags = append(tags, "pdf")
	case ".epub":
		tags = append(tags, "epub")
	case ".djvu":
		tags = append(tags, "djvu")
	case ".jpg", ".png", ".gif", ".bmp":
		tags = append(tags, "plate", "scan")
	case ".txt":
		tags = append(tags, "manuscript", "text")
	}

	return dedupeStrings(tags), meta
}

func dedupeStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// EnhanceEntry applies category-specific metadata and tagging to an IndexEntry.
func EnhanceEntry(entry *db.IndexEntry) {
	switch entry.Category {
	case "books":
		tags, meta := ParseBookMetadata(entry.Path, entry.FileName, entry.SizeBytes)
		entry.Tags = tags
		entry.Metadata = meta
	case "audio":
		entry.Tags = append(entry.Tags, "audio", strings.TrimPrefix(entry.Extension, "."))
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]any)
		}
		entry.Metadata["type"] = "audio_recording"
	case "video":
		entry.Tags = append(entry.Tags, "video", strings.TrimPrefix(entry.Extension, "."))
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]any)
		}
		entry.Metadata["type"] = "video_media"
	case "code":
		entry.Tags = append(entry.Tags, "code", strings.TrimPrefix(entry.Extension, "."))
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]any)
		}
		// Extract repo name from top directory under code/
		parts := strings.Split(entry.Path, "/")
		if len(parts) > 1 {
			entry.Metadata["repo"] = parts[1]
			entry.Tags = append(entry.Tags, strings.ToLower(parts[1]))
		}
	case "text":
		EnhanceTextEntry(entry, nil)
	case "photos", "images":
		EnhancePhotoEntry(entry, nil)
	}
	entry.Tags = dedupeStrings(entry.Tags)
}
