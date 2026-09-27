package axismundi

import "time"

// WorkspaceType represents the categorized type of Google Workspace item.
type WorkspaceType string

const (
	TypeKeep     WorkspaceType = "keep"
	TypeGmail    WorkspaceType = "gmail"
	TypeDoc      WorkspaceType = "doc"
	TypeSheet    WorkspaceType = "sheet"
	TypeCalendar WorkspaceType = "calendar"
)

// WorkspaceItem is the unified representation of an observed workspace artifact.
type WorkspaceItem struct {
	ID            string         `json:"id"`
	Type          WorkspaceType  `json:"type"`
	Title         string         `json:"title"`
	Snippet       string         `json:"snippet"`
	Status        string         `json:"status,omitempty"`
	Source        string         `json:"source"` // "keep", "gmail", "docs", "sheets", "mcp"
	FirstSeenAt   time.Time      `json:"first_seen_at"`
	LastSeenAt    time.Time      `json:"last_seen_at"`
	AlertEmitted  bool           `json:"alert_emitted"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// WorkspaceAlert represents a notification emitted when a new workspace item is detected.
type WorkspaceAlert struct {
	ID        string        `json:"id"`
	ItemID    string        `json:"item_id"`
	Type      WorkspaceType `json:"type"`
	Title     string        `json:"title"`
	Snippet   string        `json:"snippet"`
	Timestamp time.Time     `json:"timestamp"`
	Message   string        `json:"message"`
}

// WorkspaceCounts summarizes the counts of active items by type.
type WorkspaceCounts struct {
	Total     int `json:"total"`
	KeepNotes int `json:"keep_notes"`
	Gmail     int `json:"gmail"`
	Docs      int `json:"docs"`
	Sheets    int `json:"sheets"`
	Calendar  int `json:"calendar"`
}

// WorkspaceStatus provides diagnostics on the Axis Mundi connection and polling engine.
type WorkspaceStatus struct {
	IsLive         bool            `json:"is_live"`
	Endpoint       string          `json:"endpoint"`
	LastPolledAt   time.Time       `json:"last_polled_at"`
	LastPollStatus string          `json:"last_poll_status"`
	Mode           string          `json:"mode"`
	Counts         WorkspaceCounts `json:"counts"`
	ActiveAlerts   int             `json:"active_alerts"`
}

// WorkspaceFeed contains a collection of observed items and recent alerts.
type WorkspaceFeed struct {
	Status    WorkspaceStatus  `json:"status"`
	Alerts    []WorkspaceAlert `json:"alerts"`
	Items     []WorkspaceItem  `json:"items"`
	Total     int              `json:"total"`
	Generated time.Time        `json:"generated_at"`
}
