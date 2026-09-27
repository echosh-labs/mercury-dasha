package grecorder

import (
	"encoding/json"
	"time"
)

// RPCRequest is a standard JSON-RPC 2.0 request payload.
type RPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// RPCResponse is a standard JSON-RPC 2.0 response payload.
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a standard JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}

// ToolCallParams encapsulates parameters for tools/call.
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ToolResult represents the output of a tools/call execution.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent is a content item within a tool result.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Recording represents a Google Recorder recording metadata record.
type Recording struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	RecordedAt     time.Time `json:"recorded_at"`
	Duration       string    `json:"duration"`
	DurationMs     int64     `json:"duration_ms"`
	Location       string    `json:"location,omitempty"`
	HasTranscript  bool      `json:"has_transcript"`
	IsSynced       bool      `json:"is_synced,omitempty"`
	AudioPath      string    `json:"audio_path,omitempty"`
	TranscriptTxt  string    `json:"transcript_txt,omitempty"`
	TranscriptJSON string    `json:"transcript_json,omitempty"`
}

// ManifestEntry represents a synced recording in the local Dropbox destination folder.
type ManifestEntry struct {
	RecordingID        string    `json:"recording_id"`
	DeviceID           string    `json:"device_id,omitempty"`
	Title              string    `json:"title"`
	RecordedAt         time.Time `json:"recorded_at"`
	Duration           string    `json:"duration"`
	DurationMs         int64     `json:"duration_ms"`
	Location           string    `json:"location,omitempty"`
	AudioFile          string    `json:"audio_file,omitempty"`
	AudioSizeBytes     int64     `json:"audio_size_bytes,omitempty"`
	TranscriptTextFile string    `json:"transcript_text_file,omitempty"`
	TranscriptTextSize int64     `json:"transcript_text_size_bytes,omitempty"`
	TranscriptJSONFile string    `json:"transcript_json_file,omitempty"`
	TranscriptJSONSize int64     `json:"transcript_json_size_bytes,omitempty"`
	SyncedAt           time.Time `json:"synced_at"`
	HasTranscript      bool      `json:"has_transcript"`
}

// SyncManifest tracks synchronization state across executions in the target folder.
type SyncManifest struct {
	Version    string                   `json:"version"`
	LastSync   time.Time                `json:"last_sync"`
	Recordings map[string]ManifestEntry `json:"recordings"`
}

// SyncOptions specifies parameters for triggering a sync.
type SyncOptions struct {
	Limit              int    `json:"limit,omitempty"`
	All                bool   `json:"all,omitempty"`
	Query              string `json:"query,omitempty"`
	Force              bool   `json:"force,omitempty"`
	DownloadAudio      bool   `json:"download_audio"`
	DownloadTranscript bool   `json:"download_transcript"`
	TranscriptFormat   string `json:"transcript_format,omitempty"` // "text", "json", "both"
	DownloadMetadata   bool   `json:"download_metadata"`
}

// SyncStatus reports live progress of an ongoing or completed sync operation.
type SyncStatus struct {
	IsSyncing       bool       `json:"is_syncing"`
	CurrentItem     string     `json:"current_item,omitempty"`
	TotalSeen       int        `json:"total_seen"`
	TotalDownloaded int        `json:"total_downloaded"`
	TotalSkipped    int        `json:"total_skipped"`
	TotalErrors     int        `json:"total_errors"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
	ManifestTotal   int        `json:"manifest_total"`
}

// StatusOverview reports comprehensive health, MCP connectivity, and storage stats.
type StatusOverview struct {
	MCPConnected         bool      `json:"mcp_connected"`
	MCPURL               string    `json:"mcp_url"`
	AuthValid            bool      `json:"auth_valid"`
	OutputDir            string    `json:"output_dir"`
	TotalLocalRecordings int       `json:"total_local_recordings"`
	LastSyncTime         time.Time `json:"last_sync_time"`
	AutoSyncMins         int       `json:"auto_sync_mins"`
	IsSyncing            bool      `json:"is_syncing"`
}

// DiarizedParagraph represents a structured speaker block with millisecond timings.
type DiarizedParagraph struct {
	Speaker string `json:"speaker"`
	StartMs int64  `json:"start_ms"`
	EndMs   int64  `json:"end_ms"`
	Text    string `json:"text"`
}

// TranscriptPayload bundles both text and structured diarized segments.
type TranscriptPayload struct {
	RecordingID string              `json:"recording_id"`
	RawText     string              `json:"raw_text"`
	Paragraphs  []DiarizedParagraph `json:"paragraphs"`
	JSON        any                 `json:"json,omitempty"`
}
