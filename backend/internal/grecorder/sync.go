package grecorder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var (
	txtDiarizedRegex    = regexp.MustCompile(`^\[(\d{1,2}):(\d{2})(?::(\d{2}))?\]\s*(?:([^:]+):\s*)?(.*)$`)
	txtSpeakerTimeRegex = regexp.MustCompile(`^\[([^\]]+)\]\s*\((?:(\d{1,2}):)?(\d{1,2}):(\d{2})\)`)
)

// Service coordinates Google Recorder MCP interactions, local Dropbox syncing, and BoltDB indexing.
type Service struct {
	client       *Client
	store        db.StorageEngine
	dropboxRoot  string
	outputDir    string
	autoSyncMins int
	mu           sync.RWMutex
	status       SyncStatus
	lastSyncTime time.Time
	stopPoller   chan struct{}
}

// NewService constructs a Google Recorder synchronization service.
func NewService(cfg *config.Config, store db.StorageEngine) *Service {
	client := NewClient(cfg.RecorderMCPURL)

	s := &Service{
		client:       client,
		store:        store,
		dropboxRoot:  cfg.DropboxLocalPath,
		outputDir:    cfg.RecorderOutputDir,
		autoSyncMins: cfg.RecorderAutoSyncMins,
		stopPoller:   make(chan struct{}),
	}

	// Read initial manifest if present
	if manifest, err := s.readManifest(); err == nil {
		s.lastSyncTime = manifest.LastSync
		s.status.ManifestTotal = len(manifest.Recordings)
	}

	return s
}

// StartBackgroundPoller starts the periodic 30-minute auto-synchronization loop.
func (s *Service) StartBackgroundPoller() {
	if s.autoSyncMins <= 0 {
		log.Printf("[GRecorder] Periodic auto-sync disabled (interval=%d)", s.autoSyncMins)
		return
	}

	interval := time.Duration(s.autoSyncMins) * time.Minute
	log.Printf("[GRecorder] Starting background synchronization poller (cadence: %v)...", interval)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopPoller:
				log.Printf("[GRecorder] Stopped background synchronization poller.")
				return
			case <-ticker.C:
				log.Printf("[GRecorder] Background auto-sync triggered (%d min cycle)", s.autoSyncMins)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
				opts := SyncOptions{
					All:                true,
					DownloadAudio:      true,
					DownloadTranscript: true,
					TranscriptFormat:   "both",
					DownloadMetadata:   true,
				}
				if _, err := s.Sync(ctx, opts); err != nil {
					log.Printf("[GRecorder] Background auto-sync encountered warning: %v", err)
				}
				cancel()
			}
		}
	}()
}

// StopBackgroundPoller halts the periodic timer.
func (s *Service) StopBackgroundPoller() {
	close(s.stopPoller)
}

// Sync runs an incremental library synchronization pass via MCP server into Dropbox.
func (s *Service) Sync(ctx context.Context, opts SyncOptions) (*SyncStatus, error) {
	s.mu.Lock()
	if s.status.IsSyncing {
		s.mu.Unlock()
		return &s.status, fmt.Errorf("synchronization already in progress")
	}

	started := time.Now()
	s.status = SyncStatus{
		IsSyncing:       true,
		StartedAt:       &started,
		TotalSeen:       0,
		TotalDownloaded: 0,
		TotalSkipped:    0,
		TotalErrors:     0,
		LastError:       "",
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		now := time.Now()
		s.status.IsSyncing = false
		s.status.FinishedAt = &now
		s.lastSyncTime = now
		if manifest, err := s.readManifest(); err == nil {
			s.status.ManifestTotal = len(manifest.Recordings)
		}
		s.mu.Unlock()
	}()

	// Ensure destination directory exists in Dropbox
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		s.recordError(fmt.Sprintf("Failed to create destination dir %s: %v", s.outputDir, err))
		return &s.status, err
	}

	// Default opts
	if !opts.DownloadAudio && !opts.DownloadTranscript {
		opts.DownloadAudio = true
		opts.DownloadTranscript = true
		opts.TranscriptFormat = "both"
		opts.DownloadMetadata = true
	}

	log.Printf("[GRecorder] Executing sync_recordings via MCP (target: %s, all=%v)...", s.outputDir, opts.All)
	syncOutput, err := s.client.SyncRecordings(ctx, s.outputDir, opts)
	if err != nil {
		log.Printf("[GRecorder] MCP sync_recordings error: %v (output: %s)", err, syncOutput)
		s.recordError(err.Error())
		return &s.status, err
	}

	log.Printf("[GRecorder] MCP sync succeeded: %s", syncOutput)

	// Read updated manifest and update BoltDB index entries
	manifest, err := s.readManifest()
	if err == nil {
		s.mu.Lock()
		s.status.ManifestTotal = len(manifest.Recordings)
		s.mu.Unlock()

		if errIndex := s.IndexSyncedRecordings(manifest); errIndex != nil {
			log.Printf("[GRecorder] BoltDB index synchronization warning: %v", errIndex)
		}
	}

	return &s.status, nil
}

func (s *Service) recordError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.TotalErrors++
	s.status.LastError = msg
}

// GetStatus returns overarching status including MCP connectivity and local file counts.
func (s *Service) GetStatus(ctx context.Context) StatusOverview {
	s.mu.RLock()
	lastSync := s.lastSyncTime
	isSyncing := s.status.IsSyncing
	totalLocal := s.status.ManifestTotal
	s.mu.RUnlock()

	connected := s.client.Ping(ctx)
	authValid := false
	if connected {
		authValid, _ = s.client.AuthStatus(ctx)
	}

	return StatusOverview{
		MCPConnected:         connected,
		MCPURL:               s.client.mcpURL,
		AuthValid:            authValid,
		OutputDir:            s.outputDir,
		TotalLocalRecordings: totalLocal,
		LastSyncTime:         lastSync,
		AutoSyncMins:         s.autoSyncMins,
		IsSyncing:            isSyncing,
	}
}

// GetSyncStatus returns live progress details.
func (s *Service) GetSyncStatus() SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// ListRecordings returns recordings combining local manifest with remote MCP catalog.
func (s *Service) ListRecordings(ctx context.Context, query string) ([]Recording, error) {
	manifest, _ := s.readManifest()
	localMap := make(map[string]ManifestEntry)
	if manifest != nil && manifest.Recordings != nil {
		localMap = manifest.Recordings
	}

	var results []Recording

	// 1. If MCP is online, query remote recordings
	if s.client.Ping(ctx) {
		remoteRecs, err := s.client.ListRecordings(ctx, 100, query, true)
		if err == nil {
			for _, rec := range remoteRecs {
				if local, ok := localMap[rec.ID]; ok {
					rec.IsSynced = true
					rec.AudioPath = filepath.Join(s.outputDir, local.AudioFile)
					rec.TranscriptTxt = filepath.Join(s.outputDir, local.TranscriptTextFile)
					rec.TranscriptJSON = filepath.Join(s.outputDir, local.TranscriptJSONFile)
					rec.Duration = local.Duration
					rec.DurationMs = local.DurationMs
				}
				results = append(results, rec)
			}
			return results, nil
		}
	}

	// 2. Offline / Local fallback: populate from local manifest
	for id, entry := range localMap {
		if query != "" && !strings.Contains(strings.ToLower(entry.Title), strings.ToLower(query)) {
			continue
		}
		results = append(results, Recording{
			ID:             id,
			Title:          entry.Title,
			RecordedAt:     entry.RecordedAt,
			Duration:       entry.Duration,
			DurationMs:     entry.DurationMs,
			Location:       entry.Location,
			HasTranscript:  entry.HasTranscript,
			IsSynced:       true,
			AudioPath:      filepath.Join(s.outputDir, entry.AudioFile),
			TranscriptTxt:  filepath.Join(s.outputDir, entry.TranscriptTextFile),
			TranscriptJSON: filepath.Join(s.outputDir, entry.TranscriptJSONFile),
		})
	}

	return results, nil
}

// GetTranscript loads the transcript for a recording, preferring local disk and falling back to MCP.
func (s *Service) GetTranscript(ctx context.Context, recordingID string) (*TranscriptPayload, error) {
	manifest, _ := s.readManifest()
	var jsonPath, txtPath string

	if manifest != nil && manifest.Recordings != nil {
		if entry, ok := manifest.Recordings[recordingID]; ok {
			if entry.TranscriptJSONFile != "" {
				jsonPath = filepath.Join(s.outputDir, entry.TranscriptJSONFile)
			}
			if entry.TranscriptTextFile != "" {
				txtPath = filepath.Join(s.outputDir, entry.TranscriptTextFile)
			}
		}
	}

	payload := &TranscriptPayload{
		RecordingID: recordingID,
	}

	// 1. Try local JSON first for exact millisecond timestamps and speakers
	if jsonPath != "" {
		if data, err := os.ReadFile(jsonPath); err == nil {
			var structured struct {
				RecordingID string `json:"recording_id"`
				FullText    string `json:"full_text"`
				Segments    []struct {
					Speaker   string `json:"speaker"`
					StartTime string `json:"start_time"`
					StartMs   int64  `json:"start_ms"`
					EndTime   string `json:"end_time"`
					EndMs     int64  `json:"end_ms"`
					Text      string `json:"text"`
				} `json:"segments"`
				Transcript struct {
					RecordingID string `json:"recording_id"`
					FullText    string `json:"full_text"`
					Segments    []struct {
						Speaker   string `json:"speaker"`
						StartTime string `json:"start_time"`
						StartMs   int64  `json:"start_ms"`
						EndTime   string `json:"end_time"`
						EndMs     int64  `json:"end_ms"`
						Text      string `json:"text"`
					} `json:"segments"`
				} `json:"transcript"`
			}
			if err := json.Unmarshal(data, &structured); err == nil {
				segs := structured.Segments
				fullText := structured.FullText
				if len(segs) == 0 && len(structured.Transcript.Segments) > 0 {
					segs = structured.Transcript.Segments
					fullText = structured.Transcript.FullText
				}

				if len(segs) > 0 {
					payload.RawText = fullText
					for _, seg := range segs {
						payload.Paragraphs = append(payload.Paragraphs, DiarizedParagraph{
							Speaker: seg.Speaker,
							StartMs: seg.StartMs,
							EndMs:   seg.EndMs,
							Text:    seg.Text,
						})
					}
					_ = json.Unmarshal(data, &payload.JSON)
					return payload, nil
				}
			}
		}
	}

	// 2. Try local TXT transcript fallback
	if txtPath != "" {
		if txtData, err := os.ReadFile(txtPath); err == nil {
			raw := string(txtData)
			payload.RawText = raw
			payload.Paragraphs = parseTextTranscript(raw)
			return payload, nil
		}
	}

	// 3. Fall back to MCP direct fetch
	if s.client.Ping(ctx) {
		mcpOut, err := s.client.GetTranscript(ctx, recordingID, "json")
		if err == nil {
			var structured struct {
				FullText string `json:"full_text"`
				Segments []struct {
					Speaker string `json:"speaker"`
					StartMs int64  `json:"start_ms"`
					EndMs   int64  `json:"end_ms"`
					Text    string `json:"text"`
				} `json:"segments"`
			}
			if err := json.Unmarshal([]byte(mcpOut), &structured); err == nil {
				payload.RawText = structured.FullText
				for _, seg := range structured.Segments {
					payload.Paragraphs = append(payload.Paragraphs, DiarizedParagraph{
						Speaker: seg.Speaker,
						StartMs: seg.StartMs,
						EndMs:   seg.EndMs,
						Text:    seg.Text,
					})
				}
				_ = json.Unmarshal([]byte(mcpOut), &payload.JSON)
				return payload, nil
			}
		}

		// Text fallback from MCP
		txtOut, errTxt := s.client.GetTranscript(ctx, recordingID, "text")
		if errTxt == nil {
			payload.RawText = txtOut
			payload.Paragraphs = parseTextTranscript(txtOut)
			return payload, nil
		}
	}

	return nil, fmt.Errorf("transcript not found for recording %s", recordingID)
}

// GetAudioFilePath returns the absolute POSIX path to the local .m4a audio file.
func (s *Service) GetAudioFilePath(recordingID string) (string, error) {
	manifest, err := s.readManifest()
	if err != nil {
		return "", err
	}

	entry, ok := manifest.Recordings[recordingID]
	if !ok || entry.AudioFile == "" {
		return "", fmt.Errorf("audio recording %s is not downloaded locally", recordingID)
	}

	fullPath := filepath.Join(s.outputDir, entry.AudioFile)
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("audio file missing on disk: %s", fullPath)
	}

	return fullPath, nil
}

// IndexSyncedRecordings synchronizes manifest items into Mercury Dasha's BoltDB Storehouse index.
func (s *Service) IndexSyncedRecordings(manifest *SyncManifest) error {
	if s.store == nil || manifest == nil {
		return nil
	}

	var batch []db.IndexEntry
	now := time.Now().UTC()

	for _, entry := range manifest.Recordings {
		// 1. Audio file index entry
		if entry.AudioFile != "" {
			audioFull := filepath.Join(s.outputDir, entry.AudioFile)
			relPath := strings.TrimPrefix(filepath.ToSlash(audioFull), strings.TrimRight(filepath.ToSlash(s.dropboxRoot), "/")+"/")

			batch = append(batch, db.IndexEntry{
				ID:        relPath,
				Category:  "audio",
				Path:      relPath,
				FullPath:  audioFull,
				FileName:  entry.AudioFile,
				Extension: ".m4a",
				SizeBytes: entry.AudioSizeBytes,
				ModTime:   entry.RecordedAt,
				IndexedAt: now,
				Tags:      []string{"google-recorder", "m4a", "voice", "transcribed"},
				Metadata: map[string]any{
					"recording_id":   entry.RecordingID,
					"duration":       entry.Duration,
					"duration_ms":    entry.DurationMs,
					"has_transcript": entry.HasTranscript,
					"source":         "google-recorder-mcp",
				},
			})
		}

		// 2. Transcript file index entry
		if entry.TranscriptTextFile != "" {
			txtFull := filepath.Join(s.outputDir, entry.TranscriptTextFile)
			relPath := strings.TrimPrefix(filepath.ToSlash(txtFull), strings.TrimRight(filepath.ToSlash(s.dropboxRoot), "/")+"/")

			snippet := ""
			if data, err := os.ReadFile(txtFull); err == nil {
				s := string(data)
				if len(s) > 1024 {
					snippet = s[:1024]
				} else {
					snippet = s
				}
			}

			batch = append(batch, db.IndexEntry{
				ID:        relPath,
				Category:  "text",
				Path:      relPath,
				FullPath:  txtFull,
				FileName:  entry.TranscriptTextFile,
				Extension: ".txt",
				SizeBytes: entry.TranscriptTextSize,
				ModTime:   entry.RecordedAt,
				IndexedAt: now,
				Tags:      []string{"google-recorder", "transcript", "diarized"},
				Snippet:   snippet,
				Metadata: map[string]any{
					"recording_id": entry.RecordingID,
					"source":       "google-recorder-mcp",
				},
			})
		}
	}

	if len(batch) > 0 {
		return s.store.BatchPutIndexEntries(batch)
	}
	return nil
}

func (s *Service) readManifest() (*SyncManifest, error) {
	manifestPath := filepath.Join(s.outputDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var m SyncManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// parseTextTranscript parses timestamped plain text transcripts into structured paragraphs.
func parseTextTranscript(raw string) []DiarizedParagraph {
	lines := strings.Split(raw, "\n")
	var result []DiarizedParagraph
	var pendingSpeaker string
	var pendingStartMs int64

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check format: [Speaker 1] (00:03) or [Speaker 1] (01:02:03)
		if mSpeaker := txtSpeakerTimeRegex.FindStringSubmatch(trimmed); len(mSpeaker) > 0 {
			speaker := mSpeaker[1]
			mins, _ := strconv.ParseInt(mSpeaker[3], 10, 64)
			secs, _ := strconv.ParseInt(mSpeaker[4], 10, 64)
			totalSecs := mins*60 + secs
			if mSpeaker[2] != "" {
				hrs, _ := strconv.ParseInt(mSpeaker[2], 10, 64)
				totalSecs = hrs*3600 + mins*60 + secs
			}
			pendingSpeaker = speaker
			pendingStartMs = totalSecs * 1000
			continue
		}

		// Check format: [00:03] Speaker 1: text
		m := txtDiarizedRegex.FindStringSubmatch(trimmed)
		if len(m) > 0 {
			mins, _ := strconv.ParseInt(m[1], 10, 64)
			secs, _ := strconv.ParseInt(m[2], 10, 64)
			totalSecs := mins*60 + secs
			if m[3] != "" {
				hrs := mins
				mins = secs
				s, _ := strconv.ParseInt(m[3], 10, 64)
				totalSecs = hrs*3600 + mins*60 + s
			}

			speaker := m[4]
			if speaker == "" {
				speaker = "Speaker"
			}
			text := strings.TrimSpace(m[5])

			result = append(result, DiarizedParagraph{
				Speaker: speaker,
				StartMs: totalSecs * 1000,
				Text:    text,
			})
			pendingSpeaker = ""
		} else if pendingSpeaker != "" {
			result = append(result, DiarizedParagraph{
				Speaker: pendingSpeaker,
				StartMs: pendingStartMs,
				Text:    trimmed,
			})
			pendingSpeaker = ""
		} else {
			// Append un-timestamped line to previous paragraph if present
			if len(result) > 0 {
				result[len(result)-1].Text += " " + trimmed
			} else {
				result = append(result, DiarizedParagraph{
					Speaker: "Speaker",
					StartMs: 0,
					Text:    trimmed,
				})
			}
		}
	}

	return result
}
