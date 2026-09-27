package syncthing

import (
	"encoding/json"
	"time"
)

// Event represents a single raw event from the Syncthing Event Stream API.
type Event struct {
	ID       int64           `json:"id"`
	GlobalID int64           `json:"globalID"`
	Type     string          `json:"type"`
	Time     time.Time       `json:"time"`
	Data     json.RawMessage `json:"data"`
}

// ItemFinishedData contains payload details for the ItemFinished event.
type ItemFinishedData struct {
	Folder     string `json:"folder"`
	Item       string `json:"item"`
	Error      string `json:"error"`
	Type       string `json:"type"`
	Action     string `json:"action"`
	ModifiedBy string `json:"modifiedBy"`
}

// StateChangedData contains payload details for folder state changes.
type StateChangedData struct {
	Folder string `json:"folder"`
	From   string `json:"from"`
	To     string `json:"to"`
}

// FolderConfig represents a synchronized folder configuration from Syncthing.
type FolderConfig struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	Path           string `json:"path"`
	Type           string `json:"type"`
	RescanInterval int    `json:"rescanIntervalS"`
	Paused         bool   `json:"paused"`
	State          string `json:"state,omitempty"`
	GlobalFiles    int64  `json:"globalFiles,omitempty"`
	GlobalBytes    int64  `json:"globalBytes,omitempty"`
}

// DeviceConfig represents a paired remote device in Syncthing.
type DeviceConfig struct {
	DeviceID  string    `json:"deviceID"`
	Name      string    `json:"name"`
	Address   string    `json:"address,omitempty"`
	Paused    bool      `json:"paused"`
	Connected bool      `json:"connected"`
	LastSeen  time.Time `json:"lastSeen,omitempty"`
}

// FolderDBStatus holds statistical metrics for a single sync folder.
type FolderDBStatus struct {
	GlobalFiles int64  `json:"globalFiles"`
	GlobalBytes int64  `json:"globalBytes"`
	LocalFiles  int64  `json:"localFiles"`
	LocalBytes  int64  `json:"localBytes"`
	State       string `json:"state"`
}

// RecentMediaItem represents an enriched photo or video synchronized from mobile.
type RecentMediaItem struct {
	ID              string         `json:"id"`
	FileName        string         `json:"file_name"`
	Category        string         `json:"category"`
	Path            string         `json:"path"`
	FullPath        string         `json:"full_path"`
	SizeBytes       int64          `json:"size_bytes"`
	ModTime         time.Time      `json:"mod_time"`
	DurationSec     float64        `json:"duration_sec,omitempty"`
	Resolution      string         `json:"resolution,omitempty"`
	ThumbnailURL    string         `json:"thumbnail_url,omitempty"`
	StreamURL       string         `json:"stream_url"`
	Tags            []string       `json:"tags"`
	DashaMahadasha  string         `json:"dasha_mahadasha,omitempty"`
	DashaAntardasha string         `json:"dasha_antardasha,omitempty"`
	SacredMetal     string         `json:"sacred_metal,omitempty"`
	Hora            string         `json:"hora,omitempty"`
	HermeticAxiom   string         `json:"hermetic_axiom,omitempty"`
	DeviceModel     string         `json:"device_model,omitempty"`
	Location        string         `json:"location,omitempty"`
	SyncedAt        time.Time      `json:"synced_at"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// StatusResponse encapsulates complete Syncthing subsystem status for Mercury Dasha.
type StatusResponse struct {
	SyncthingLive   bool              `json:"syncthing_live"`
	SyncthingURL    string            `json:"syncthing_url"`
	Version         string            `json:"version,omitempty"`
	MediaRoot       string            `json:"media_root"`
	TotalVaultFiles int               `json:"total_vault_files"`
	TotalVaultBytes int64             `json:"total_vault_bytes"`
	Devices         []DeviceConfig    `json:"devices"`
	Folders         []FolderConfig    `json:"folders"`
	LastEventID     int64             `json:"last_event_id"`
	IsSyncing       bool              `json:"is_syncing"`
	RecentItems     []RecentMediaItem `json:"recent_items,omitempty"`
}
