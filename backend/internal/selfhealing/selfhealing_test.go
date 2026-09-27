package selfhealing

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func createTestStore(t *testing.T) (*db.Store, func()) {
	tmpDir, err := os.MkdirTemp("", "mercury-selfhealing-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to open boltdb: %v", err)
	}
	cleanup := func() {
		_ = store.Close()
		_ = os.RemoveAll(tmpDir)
	}
	return store, cleanup
}

func TestRestartManager(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	rm := NewRestartManager(store, "secret-token-123")

	// 1. Token validation
	if !rm.ValidateToken("secret-token-123") {
		t.Errorf("expected valid token to pass")
	}
	if rm.ValidateToken("wrong-token") {
		t.Errorf("expected invalid token to fail")
	}
	if rm.ValidateToken("") {
		t.Errorf("expected empty token to fail")
	}

	// Unauthenticated when token is empty in config
	rmOpen := NewRestartManager(store, "")
	if !rmOpen.ValidateToken("") {
		t.Errorf("expected open mode to permit any token")
	}

	// 2. Trigger restart
	req := RestartRequest{
		Reason:    "Unit test restart",
		Initiator: "test_runner",
		Delay:     10 * time.Millisecond,
		DryRun:    true,
	}

	if err := rm.TriggerRestart(req); err != nil {
		t.Fatalf("failed to trigger restart: %v", err)
	}

	if !rm.IsRestarting() {
		t.Errorf("expected IsRestarting to be true")
	}

	// Second concurrent restart should be rejected
	err := rm.TriggerRestart(req)
	if err != ErrRestartInProgress {
		t.Errorf("expected ErrRestartInProgress, got %v", err)
	}

	// Read from channel
	select {
	case received := <-rm.RestartChan():
		if received.Reason != "Unit test restart" {
			t.Errorf("unexpected reason: %s", received.Reason)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for restart channel")
	}

	// Check history
	hist := rm.GetHistory()
	if len(hist) == 0 {
		t.Fatalf("expected restart history to contain at least 1 record")
	}
	if hist[0].Reason != "Unit test restart" {
		t.Errorf("history record reason mismatch: %s", hist[0].Reason)
	}

	// Test DryRun PerformReexec
	rm.PerformReexec(RestartRequest{DryRun: true})
	if rm.IsRestarting() {
		t.Errorf("expected DryRun to clear isRestarting")
	}
}

func TestControllerHealthEvaluation(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	cfg := &config.Config{
		ServiceName:         "mercury-dasha-test",
		Environment:         "test",
		Port:                "19876",
		MemoryLimitMB:       1024,
		WatchdogIntervalSec: 1,
	}
	rm := NewRestartManager(store, "")
	ctrl := NewController(store, cfg, rm)

	report := ctrl.EvaluateHealth()
	if report.OverallStatus != StatusHealthy {
		t.Errorf("expected initial status to be HEALTHY, got %v", report.OverallStatus)
	}

	if report.Components["database"].Status != StatusHealthy {
		t.Errorf("expected database to be HEALTHY, got %v", report.Components["database"].Status)
	}

	if report.Components["memory"].Status != StatusHealthy {
		t.Errorf("expected memory to be HEALTHY, got %v", report.Components["memory"].Status)
	}
}

func TestControllerRemediations(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	cfg := &config.Config{
		ServiceName:   "mercury-dasha-test",
		Environment:   "test",
		Port:          "19876",
		MemoryLimitMB: 1024,
	}
	rm := NewRestartManager(store, "")
	ctrl := NewController(store, cfg, rm)

	// Test memory GC remediation
	resp := ctrl.Remediate("memory_gc")
	if !resp.Success {
		t.Errorf("expected memory_gc remediation to succeed, got %v", resp.Message)
	}

	// Test DB verify remediation
	respDB := ctrl.Remediate("db_verify")
	if !respDB.Success {
		t.Errorf("expected db_verify remediation to succeed, got %v", respDB.Message)
	}

	// Test full health check remediation
	respFull := ctrl.Remediate("full_health_check")
	if !respFull.Success {
		t.Errorf("expected full_health_check remediation to succeed, got %v", respFull.Message)
	}

	// Verify incidents logged
	incidents := ctrl.GetIncidents(10)
	if len(incidents) == 0 {
		t.Errorf("expected incidents to be logged for remediations")
	}
}

func TestControllerPanicRecovery(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	cfg := &config.Config{
		ServiceName: "mercury-dasha-test",
		Environment: "test",
		Port:        "19876",
	}
	rm := NewRestartManager(store, "")
	ctrl := NewController(store, cfg, rm)

	// Simulate panic interception
	ctrl.RecordPanic("nil pointer dereference", []byte("goroutine 1 [running]:\nmain.test()"), "/api/test/panic")

	incidents := ctrl.GetIncidents(5)
	if len(incidents) == 0 {
		t.Fatalf("expected panic incident to be recorded")
	}

	if incidents[0].Component != "http_panic" {
		t.Errorf("expected component http_panic, got %s", incidents[0].Component)
	}
	if incidents[0].Severity != "critical" {
		t.Errorf("expected severity critical, got %s", incidents[0].Severity)
	}
}

func TestControllerWatchdogLifecycle(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	cfg := &config.Config{
		ServiceName:         "mercury-dasha-test",
		Environment:         "test",
		Port:                "19876",
		WatchdogIntervalSec: 1,
	}
	rm := NewRestartManager(store, "")
	ctrl := NewController(store, cfg, rm)

	ctrl.Start()
	// Duplicate start should be no-op
	ctrl.Start()

	// Wait briefly for at least one evaluation
	time.Sleep(100 * time.Millisecond)

	report := ctrl.CheckAndHeal()
	if report.OverallStatus != StatusHealthy {
		t.Errorf("expected HEALTHY report, got %s", report.OverallStatus)
	}

	ctrl.Stop()
	// Duplicate stop should be no-op
	ctrl.Stop()
}
