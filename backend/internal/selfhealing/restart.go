package selfhealing

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var (
	ErrRestartInProgress = errors.New("restart already in progress")
	ErrUnauthorized       = errors.New("unauthorized: invalid or missing restart token")
)

// RestartManager coordinates remote and automatic service restarts, token security,
// restart audit logging, and cross-platform process re-execution.
type RestartManager struct {
	mu           sync.RWMutex
	store        db.StorageEngine
	token        string
	restartChan  chan RestartRequest
	isRestarting atomic.Bool
	history      []RestartRecord
}

// NewRestartManager constructs a new RestartManager.
func NewRestartManager(store db.StorageEngine, token string) *RestartManager {
	rm := &RestartManager{
		store:       store,
		token:       token,
		restartChan: make(chan RestartRequest, 1),
		history:     make([]RestartRecord, 0),
	}
	rm.loadHistory()
	return rm
}

// ValidateToken checks if the provided token matches the configured secret.
// If no token is configured (e.g. local dev / open mode), access is permitted.
func (rm *RestartManager) ValidateToken(provided string) bool {
	if rm.token == "" {
		return true
	}
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(rm.token), []byte(provided)) == 1
}

// IsRestarting reports whether a restart is actively pending or executing.
func (rm *RestartManager) IsRestarting() bool {
	return rm.isRestarting.Load()
}

// RestartChan exposes the notification channel for server lifecycle control.
func (rm *RestartManager) RestartChan() <-chan RestartRequest {
	return rm.restartChan
}

// TriggerRestart registers a request to restart the service.
func (rm *RestartManager) TriggerRestart(req RestartRequest) error {
	if rm.isRestarting.Swap(true) && !req.Force {
		return ErrRestartInProgress
	}

	if req.Reason == "" {
		req.Reason = "Service restart requested via control channel"
	}
	if req.Initiator == "" {
		req.Initiator = "system"
	}
	if req.Delay <= 0 {
		req.Delay = 500 * time.Millisecond
	}

	record := RestartRecord{
		ID:        fmt.Sprintf("rst-%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC(),
		Reason:    req.Reason,
		Initiator: req.Initiator,
		Type:      req.Initiator,
		PID:       os.Getpid(),
		Success:   true,
		Details:   fmt.Sprintf("Scheduled delay: %v, Force: %v, DryRun: %v", req.Delay, req.Force, req.DryRun),
	}

	rm.saveRecord(record)

	// Send to channel non-blocking
	select {
	case rm.restartChan <- req:
	default:
		log.Printf("Restart request already queued on channel")
	}

	return nil
}

// CancelRestart resets the restarting flag (primarily for tests or abort scenarios).
func (rm *RestartManager) CancelRestart() {
	rm.isRestarting.Store(false)
}

// GetHistory returns an immutable copy of recent restart records (newest first).
func (rm *RestartManager) GetHistory() []RestartRecord {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	out := make([]RestartRecord, len(rm.history))
	for i, r := range rm.history {
		out[len(rm.history)-1-i] = r // Reverse order: newest first
	}
	return out
}

func (rm *RestartManager) saveRecord(rec RestartRecord) {
	rm.mu.Lock()
	rm.history = append(rm.history, rec)
	if len(rm.history) > 100 {
		rm.history = rm.history[len(rm.history)-100:]
	}
	rm.mu.Unlock()

	if rm.store != nil {
		data, err := json.Marshal(rec)
		if err == nil {
			key := fmt.Sprintf("restart:%d:%s", rec.Timestamp.UnixNano(), rec.ID)
			_ = rm.store.PutJSON(db.BucketSelfHealing, key, data)
		}
	}
}

func (rm *RestartManager) loadHistory() {
	if rm.store == nil {
		return
	}
	keys, err := rm.store.ListKeys(db.BucketSelfHealing, "restart:")
	if err != nil {
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, k := range keys {
		data, err := rm.store.GetJSON(db.BucketSelfHealing, k)
		if err != nil || len(data) == 0 {
			continue
		}
		var rec RestartRecord
		if err := json.Unmarshal(data, &rec); err == nil {
			rm.history = append(rm.history, rec)
		}
	}
}

// PerformReexec replaces the currently running process image with a fresh executable invocation.
// On POSIX platforms, syscall.Exec preserves the PID and environment in place.
// On non-POSIX or when syscall.Exec fails, a detached child process is spawned before exit.
func (rm *RestartManager) PerformReexec(req RestartRequest) {
	if req.DryRun {
		log.Printf("[Self-Healing] Dry-run restart requested. Skipping binary re-exec.")
		rm.isRestarting.Store(false)
		return
	}

	executable, err := os.Executable()
	if err != nil {
		log.Printf("❌ Failed to discover executable path: %v. Attempting os.Args[0] fallback...", err)
		executable = os.Args[0]
	}

	// Resolve symlinks to guarantee absolute sovereign binary target
	if resolved, err := filepath.EvalSymlinks(executable); err == nil && resolved != "" {
		executable = resolved
	}

	log.Printf("🔄 [Self-Healing Engine] Performing sovereign re-exec: %s with args %v", executable, os.Args)

	// Ensure any pending stdout/stderr logs are flushed
	_ = os.Stdout.Sync()
	_ = os.Stderr.Sync()

	// Platform execution strategy
	if runtime.GOOS != "windows" {
		// POSIX in-place image replacement
		execErr := syscall.Exec(executable, os.Args, os.Environ())
		if execErr != nil {
			log.Printf("⚠️ syscall.Exec returned error: %v. Falling back to background process spawn...", execErr)
		} else {
			// syscall.Exec never returns on success
			return
		}
	}

	// Fallback / Windows strategy: spawn detached process and exit cleanly
	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatalf("❌ Fatal: Failed to spawn replacement process on restart: %v", err)
	}

	log.Printf("✔ Replacement process spawned with PID %d. Exiting current process cleanly.", cmd.Process.Pid)
	os.Exit(0)
}
