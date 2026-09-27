package axismundi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/chrono"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var (
	bucketWorkspaceItems = []byte("axis_mundi_workspace")
)

// Listener manages continuous background monitoring and alerting for Axis Mundi items.
type Listener struct {
	client       *Client
	store        db.StorageEngine
	metronome    *chrono.Metronome
	mu           sync.RWMutex
	items        map[string]WorkspaceItem
	alerts       []WorkspaceAlert
	lastPolledAt time.Time
	lastPollErr  string
	lastMode     string
	isLive       bool
	stopChan     chan struct{}
	stopOnce     sync.Once
	closed       atomic.Bool
}

// NewListener creates a new Axis Mundi workspace listener.
func NewListener(baseURL, apiKey string, store db.StorageEngine, metronome *chrono.Metronome) *Listener {
	l := &Listener{
		client:    NewClient(baseURL, apiKey),
		store:     store,
		metronome: metronome,
		items:     make(map[string]WorkspaceItem),
		alerts:    make([]WorkspaceAlert, 0),
		stopChan:  make(chan struct{}),
	}
	l.loadPersistedItems()
	return l
}

func (l *Listener) isClosed() bool {
	return l.closed.Load()
}

// Start launches the background polling loop.
func (l *Listener) Start(interval time.Duration) {
	if interval <= 0 {
		interval = 20 * time.Second
	}

	// Initial poll on startup
	go func() {
		select {
		case <-l.stopChan:
			return
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = l.Poll(ctx)
	}()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-l.stopChan:
				return
			case <-ticker.C:
				if l.isClosed() {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				_, _ = l.Poll(ctx)
				cancel()
			}
		}
	}()
	log.Printf("🔭 [AxisMundi] Workspace listener active (polling %s at %v)", l.client.baseURL, interval)
}

// Stop signals the listener loop to terminate.
func (l *Listener) Stop() {
	l.stopOnce.Do(func() {
		l.closed.Store(true)
		close(l.stopChan)
	})
}

// Poll fetches current items from Axis Mundi and identifies newly arrived items.
func (l *Listener) Poll(ctx context.Context) ([]WorkspaceAlert, error) {
	live, mode, err := l.client.Ping(ctx)
	l.mu.Lock()
	l.lastPolledAt = time.Now()
	l.isLive = live
	l.lastMode = mode
	if err != nil {
		l.lastPollErr = err.Error()
		l.mu.Unlock()
		return nil, err
	}
	l.lastPollErr = ""
	l.mu.Unlock()

	items, err := l.client.FetchRegistry(ctx)
	if err != nil {
		l.mu.Lock()
		l.lastPollErr = err.Error()
		l.mu.Unlock()
		return nil, err
	}

	return l.ProcessItems(items), nil
}

// ProcessItems compares discovered items with known items and generates alerts for newly discovered items.
func (l *Listener) ProcessItems(incoming []WorkspaceItem) []WorkspaceAlert {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	var newAlerts []WorkspaceAlert

	for _, item := range incoming {
		existing, found := l.items[item.ID]
		if !found {
			// Brand new item discovered!
			item.FirstSeenAt = now
			item.LastSeenAt = now
			item.AlertEmitted = true
			l.items[item.ID] = item

			alert := WorkspaceAlert{
				ID:        fmt.Sprintf("alert-%s-%d", item.ID, now.UnixNano()),
				ItemID:    item.ID,
				Type:      item.Type,
				Title:     item.Title,
				Snippet:   item.Snippet,
				Timestamp: now,
				Message:   formatAlertMessage(item),
			}
			newAlerts = append(newAlerts, alert)
			l.alerts = append(l.alerts, alert)
			if len(l.alerts) > 100 {
				l.alerts = l.alerts[len(l.alerts)-100:] // Keep latest 100 alerts in memory
			}

			log.Printf("🔔 [AxisMundi ALERT] New %s detected: '%s' (ID: %s)", item.Type, item.Title, item.ID)

			// Persist item to BoltDB
			l.persistItem(item)

			// Publish to Chrono-Pulse SSE stream if metronome is wired
			if l.metronome != nil {
				l.metronome.BroadcastEvent("axis_mundi_alert", alert)
			}
		} else {
			// Update last seen
			existing.LastSeenAt = now
			if item.Status != "" {
				existing.Status = item.Status
			}
			if item.Title != "" {
				existing.Title = item.Title
			}
			l.items[item.ID] = existing
		}
	}

	return newAlerts
}

// IngestEvent accepts a pushed event from an external webhook or agent.
func (l *Listener) IngestEvent(item WorkspaceItem) (WorkspaceAlert, bool) {
	alerts := l.ProcessItems([]WorkspaceItem{item})
	if len(alerts) > 0 {
		return alerts[0], true
	}
	return WorkspaceAlert{}, false
}

// GetStatus returns operational diagnostics.
func (l *Listener) GetStatus() WorkspaceStatus {
	l.mu.RLock()
	defer l.mu.RUnlock()

	counts := WorkspaceCounts{}
	for _, it := range l.items {
		counts.Total++
		switch it.Type {
		case TypeKeep:
			counts.KeepNotes++
		case TypeGmail:
			counts.Gmail++
		case TypeDoc:
			counts.Docs++
		case TypeSheet:
			counts.Sheets++
		case TypeCalendar:
			counts.Calendar++
		}
	}

	statusStr := "healthy"
	if l.lastPollErr != "" {
		statusStr = "error: " + l.lastPollErr
	}

	return WorkspaceStatus{
		IsLive:         l.isLive,
		Endpoint:       l.client.baseURL,
		LastPolledAt:   l.lastPolledAt,
		LastPollStatus: statusStr,
		Mode:           l.lastMode,
		Counts:         counts,
		ActiveAlerts:   len(l.alerts),
	}
}

// GetFeed compiles the full feed of observed items.
func (l *Listener) GetFeed(filterType string, newOnly bool) WorkspaceFeed {
	l.mu.RLock()
	defer l.mu.RUnlock()

	status := l.GetStatus()
	var filteredItems []WorkspaceItem

	for _, item := range l.items {
		if filterType != "" && string(item.Type) != filterType {
			continue
		}
		if newOnly && !item.AlertEmitted {
			continue
		}
		filteredItems = append(filteredItems, item)
	}

	return WorkspaceFeed{
		Status:    status,
		Alerts:    l.alerts,
		Items:     filteredItems,
		Total:     len(filteredItems),
		Generated: time.Now(),
	}
}

func formatAlertMessage(item WorkspaceItem) string {
	switch item.Type {
	case TypeKeep:
		return fmt.Sprintf("New Google Keep Note: %s", item.Title)
	case TypeGmail:
		return fmt.Sprintf("New Gmail Thread: %s", item.Title)
	case TypeDoc:
		return fmt.Sprintf("New Google Doc: %s", item.Title)
	case TypeSheet:
		return fmt.Sprintf("New Google Sheet: %s", item.Title)
	default:
		return fmt.Sprintf("New %s Item: %s", item.Type, item.Title)
	}
}

func (l *Listener) persistItem(item WorkspaceItem) {
	if l.isClosed() || l.store == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	data, err := json.Marshal(item)
	if err != nil {
		return
	}
	_ = l.store.PutJSON(bucketWorkspaceItems, item.ID, data)
}

func (l *Listener) loadPersistedItems() {
	if l.isClosed() || l.store == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	keys, err := l.store.ListKeys(bucketWorkspaceItems, "")
	if err != nil {
		return
	}
	for _, k := range keys {
		raw, err := l.store.GetJSON(bucketWorkspaceItems, k)
		if err != nil || len(raw) == 0 {
			continue
		}
		var it WorkspaceItem
		if err := json.Unmarshal(raw, &it); err == nil {
			l.items[it.ID] = it
		}
	}
	if len(l.items) > 0 {
		log.Printf("📂 [AxisMundi] Restored %d persisted workspace items from BoltDB", len(l.items))
	}
}
