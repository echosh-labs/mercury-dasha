package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var (
	BucketSessions    = []byte("dnd_sessions")
	BucketAudioSlices = []byte("audio_slices")
)

type SessionStore struct {
	storage    db.StorageEngine
	basePath   string
	mu         sync.RWMutex
}

func NewSessionStore(storage db.StorageEngine, basePath string) *SessionStore {
	if basePath == "" {
		basePath = "/home/justin/Dropbox/MercuryDasha/sessions/dnd"
	}
	_ = os.MkdirAll(basePath, 0755)

	return &SessionStore{
		storage:  storage,
		basePath: basePath,
	}
}

// GetBasePath returns the configured base directory for audio sessions.
func (s *SessionStore) GetBasePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.basePath
}

// CreateSession initializes a new D&D audio session record and assigns its storage file path.
func (s *SessionStore) CreateSession(sess *AudioSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess.EnsureDefaults()

	// Assign storage filename
	ext := ".wav"
	if strings.Contains(sess.Format, "webm") {
		ext = ".webm"
	} else if strings.Contains(sess.Format, "flac") {
		ext = ".flac"
	} else if strings.Contains(sess.Format, "mp3") {
		ext = ".mp3"
	}

	datePrefix := sess.StartTime.Format("2006-01-02")
	cleanTitle := SanitizeFilename(sess.Title)
	filename := fmt.Sprintf("%s_%s_%s%s", datePrefix, SanitizeFilename(sess.Campaign), cleanTitle, ext)

	sess.FilePath = filepath.Join("MercuryDasha", "sessions", "dnd", filename)
	fullDiskPath := filepath.Join(s.basePath, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullDiskPath), 0755); err != nil {
		return fmt.Errorf("failed to create audio directory: %w", err)
	}

	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	if err := s.storage.PutJSON(BucketSessions, sess.ID, data); err != nil {
		return fmt.Errorf("failed to persist session in db: %w", err)
	}

	s.indexSession(sess, fullDiskPath)
	return nil
}

// GetSession retrieves an audio session by ID.
func (s *SessionStore) GetSession(id string) (*AudioSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.storage.GetJSON(BucketSessions, id)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var sess AudioSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}

	return &sess, nil
}

// ListSessions retrieves all recorded sessions, optionally filtered by campaign.
func (s *SessionStore) ListSessions(campaign string) ([]AudioSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys, err := s.storage.ListKeys(BucketSessions, "")
	if err != nil {
		return nil, err
	}

	sessions := make([]AudioSession, 0, len(keys))
	for _, key := range keys {
		data, err := s.storage.GetJSON(BucketSessions, key)
		if err != nil {
			continue
		}
		var sess AudioSession
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}
		if campaign != "" && !strings.EqualFold(sess.Campaign, campaign) {
			continue
		}
		sessions = append(sessions, sess)
	}

	// Sort newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions, nil
}

// AddMarker appends an event marker to an existing session.
func (s *SessionStore) AddMarker(sessionID string, marker SessionMarker) (*AudioSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.storage.GetJSON(BucketSessions, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	var sess AudioSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}

	if marker.ID == "" {
		marker.ID = fmt.Sprintf("m-%d", time.Now().UnixNano())
	}
	if marker.WallTime.IsZero() {
		marker.WallTime = time.Now()
	}
	if marker.TimestampMs <= 0 && !sess.StartTime.IsZero() {
		marker.TimestampMs = time.Since(sess.StartTime).Milliseconds()
	}
	if marker.Category == "" {
		marker.Category = "general"
	}

	sess.Markers = append(sess.Markers, marker)
	sess.UpdatedAt = time.Now()

	updatedData, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	if err := s.storage.PutJSON(BucketSessions, sess.ID, updatedData); err != nil {
		return nil, err
	}

	return &sess, nil
}

// AddSlice registers a newly extracted slice to the session.
func (s *SessionStore) AddSlice(sessionID string, slice AudioSlice) (*AudioSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.storage.GetJSON(BucketSessions, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	var sess AudioSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}

	if slice.ID == "" {
		slice.ID = fmt.Sprintf("slice-%d-%d", slice.StartMs, slice.EndMs)
	}
	if slice.CreatedAt.IsZero() {
		slice.CreatedAt = time.Now()
	}

	sess.Slices = append(sess.Slices, slice)
	sess.UpdatedAt = time.Now()

	updatedData, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	if err := s.storage.PutJSON(BucketSessions, sess.ID, updatedData); err != nil {
		return nil, err
	}

	// Also store slice in dedicated slice bucket
	sliceData, _ := json.Marshal(slice)
	_ = s.storage.PutJSON(BucketAudioSlices, slice.ID, sliceData)

	return &sess, nil
}

// CompleteSession finalizes a session: updates end time, duration, file size, and status.
func (s *SessionStore) CompleteSession(sessionID string) (*AudioSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.storage.GetJSON(BucketSessions, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	var sess AudioSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}

	sess.Status = "completed"
	sess.EndTime = time.Now()
	sess.UpdatedAt = time.Now()

	fullDiskPath := s.ResolveDiskPath(sess.FilePath)
	if fi, err := os.Stat(fullDiskPath); err == nil {
		sess.SizeBytes = fi.Size()

		// For WAV files, update header sizes and calculate exact duration
		if strings.HasSuffix(strings.ToLower(sess.FilePath), ".wav") {
			_ = UpdateWavHeaderSizes(fullDiskPath)
			if f, err := os.Open(fullDiskPath); err == nil {
				if info, err := ParseWavHeader(f, sess.SizeBytes); err == nil {
					sess.DurationSec = info.DurationSec
					sess.SampleRate = info.SampleRate
					sess.Channels = info.Channels
					sess.BitDepth = info.BitsPerSample
				}
				f.Close()
			}
		}
	}

	if sess.DurationSec <= 0 && !sess.StartTime.IsZero() {
		sess.DurationSec = sess.EndTime.Sub(sess.StartTime).Seconds()
	}

	updatedData, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	if err := s.storage.PutJSON(BucketSessions, sess.ID, updatedData); err != nil {
		return nil, err
	}

	s.indexSession(&sess, fullDiskPath)
	return &sess, nil
}

// DeleteSession removes session record from DB and deletes local audio file if no other session uses it.
func (s *SessionStore) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.storage.GetJSON(BucketSessions, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}

	var sess AudioSession
	_ = json.Unmarshal(data, &sess)

	// Check if another session references this same audio file
	fullDiskPath := s.ResolveDiskPath(sess.FilePath)
	otherUsing := false
	keys, _ := s.storage.ListKeys(BucketSessions, "")
	for _, key := range keys {
		if key == sessionID {
			continue
		}
		raw, err := s.storage.GetJSON(BucketSessions, key)
		if err != nil {
			continue
		}
		var other AudioSession
		if json.Unmarshal(raw, &other) == nil {
			if other.FilePath == sess.FilePath {
				otherUsing = true
				break
			}
		}
	}
	if !otherUsing && fullDiskPath != "" {
		_ = os.Remove(fullDiskPath)
	}

	return s.storage.Delete(BucketSessions, sessionID)
}

// ResolveDiskPath turns a relative storage path into an absolute disk path.
func (s *SessionStore) ResolveDiskPath(relPath string) string {
	if relPath == "" {
		return ""
	}
	if filepath.IsAbs(relPath) {
		return relPath
	}
	clean := filepath.Base(relPath)
	return filepath.Join(s.basePath, clean)
}

// indexSession pushes an index entry to the general storehouse indexer.
func (s *SessionStore) indexSession(sess *AudioSession, fullDiskPath string) {
	entry := db.IndexEntry{
		ID:        sess.FilePath,
		Category:  "audio",
		Path:      sess.FilePath,
		FullPath:  fullDiskPath,
		FileName:  filepath.Base(sess.FilePath),
		Extension: filepath.Ext(sess.FilePath),
		SizeBytes: sess.SizeBytes,
		ModTime:   sess.UpdatedAt,
		IndexedAt: time.Now(),
		Tags:      append([]string{"dnd", "session", sess.Campaign}, sess.Tags...),
		Snippet:   fmt.Sprintf("D&D Session: %s | Campaign: %s | Duration: %.1fs", sess.Title, sess.Campaign, sess.DurationSec),
		Metadata: map[string]any{
			"title":        sess.Title,
			"duration_sec": sess.DurationSec,
			"session_id":   sess.ID,
			"campaign":     sess.Campaign,
			"dm":           sess.DM,
			"markers":      len(sess.Markers),
			"sample_rate":  sess.SampleRate,
			"channels":     sess.Channels,
			"bit_depth":    sess.BitDepth,
		},
	}
	_ = s.storage.BatchPutIndexEntries([]db.IndexEntry{entry})
}
