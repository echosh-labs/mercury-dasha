package amra_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/amra"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func setupTestAmraEngine(t *testing.T) (*amra.Engine, *amra.Handler, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test_amra.db"

	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	mockProvider := amra.NewMockPaymentProvider("test_secret_123")
	_ = db.SeedEsotericContent(store)
	engine := amra.NewEngine(store, mockProvider)
	handler := amra.NewHandler(engine)

	cleanup := func() {
		_ = store.Close()
	}

	return engine, handler, cleanup
}

func TestAMRA_Plans(t *testing.T) {
	_, handler, cleanup := setupTestAmraEngine(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/amra/plans", nil)
	rec := httptest.NewRecorder()

	handler.ListPlansHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from ListPlansHandler, got %d", rec.Code)
	}

	var resp struct {
		Plans []amra.Plan `json:"plans"`
		Count int         `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Count < 4 {
		t.Errorf("expected at least 4 default plans, got %d", resp.Count)
	}
}

func TestAMRA_CheckoutSession(t *testing.T) {
	_, handler, cleanup := setupTestAmraEngine(t)
	defer cleanup()

	reqBody := amra.CheckoutRequest{
		PlanID:        "plan_magus_monthly",
		CustomerID:    "cust_justin_001",
		CustomerEmail: "justin@echosh-labs.com",
		SuccessURL:    "https://echosh-labs.com/success",
		CancelURL:     "https://echosh-labs.com/cancel",
		Provider:      "mock",
	}
	payload, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/amra/checkout", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	handler.CheckoutHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from CheckoutHandler, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp amra.CheckoutResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse checkout response: %v", err)
	}

	if resp.SessionID == "" || resp.CheckoutURL == "" {
		t.Errorf("invalid checkout response: %+v", resp)
	}
}

func TestAMRA_WebhookAndLedger(t *testing.T) {
	engine, handler, cleanup := setupTestAmraEngine(t)
	defer cleanup()

	webhookPayload := []byte(`{
		"customer_id": "cust_magus_999",
		"customer_email": "magus@echosh-labs.com",
		"plan_id": "plan_magus_monthly",
		"amount_cents": 8900,
		"currency": "USD",
		"subscription_id": "sub_magus_999"
	}`)

	// 1. Process first webhook event
	tx, err := engine.ProcessWebhookIngest(context.Background(), "mock", "checkout.session.completed", webhookPayload)
	if err != nil {
		t.Fatalf("unexpected error processing webhook: %v", err)
	}
	if tx.AmountCents != 8900 {
		t.Errorf("expected transaction amount 8900, got %d", tx.AmountCents)
	}

	// 2. Test Idempotency: Re-processing identical webhook must return ErrDuplicateEvent
	_, errDuplicate := engine.ProcessWebhookIngest(context.Background(), "mock", "checkout.session.completed", webhookPayload)
	if errDuplicate != amra.ErrDuplicateEvent {
		t.Fatalf("expected ErrDuplicateEvent on duplicate webhook, got %v", errDuplicate)
	}

	// 3. Verify Subscription in BoltDB
	sub, err := engine.GetSubscription("sub_magus_999")
	if err != nil {
		t.Fatalf("failed to retrieve subscription: %v", err)
	}
	if sub.Status != amra.StatusActive {
		t.Errorf("expected subscription status 'active', got %s", sub.Status)
	}

	// 4. Query Ledger
	ledgerReq := httptest.NewRequest(http.MethodGet, "/api/v1/amra/ledger?limit=10", nil)
	ledgerRec := httptest.NewRecorder()
	handler.LedgerHandler(ledgerRec, ledgerReq)

	if ledgerRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from LedgerHandler, got %d", ledgerRec.Code)
	}

	var ledgerResp struct {
		Ledger []amra.Transaction `json:"ledger"`
		Count  int                `json:"count"`
	}
	if err := json.Unmarshal(ledgerRec.Body.Bytes(), &ledgerResp); err != nil {
		t.Fatalf("failed to unmarshal ledger response: %v", err)
	}

	if ledgerResp.Count != 1 {
		t.Fatalf("expected 1 ledger transaction, got %d", ledgerResp.Count)
	}
	if ledgerResp.Ledger[0].CustomerID != "cust_magus_999" {
		t.Errorf("expected customer cust_magus_999, got %s", ledgerResp.Ledger[0].CustomerID)
	}
}

func TestAMRA_YouTubeBridgeAndMetrics(t *testing.T) {
	engine, handler, cleanup := setupTestAmraEngine(t)
	defer cleanup()

	ctx := context.Background()

	// Ingest YouTube revenue record
	rec := amra.YouTubeRevenueRecord{
		Day:              "2026-09-09",
		EstimatedRevenue: 48.50,
		MonetizedPlays:   8500,
		CPM:              5.70,
		ChannelID:        "UC_test_channel_123",
	}

	tx, err := engine.IngestYouTubeRevenue(ctx, rec)
	if err != nil {
		t.Fatalf("failed to ingest youtube revenue: %v", err)
	}
	if tx.AmountCents != 4850 || tx.Provider != "youtube_partner" {
		t.Errorf("unexpected tx: %+v", tx)
	}

	// Ingest duplicate -> expect ErrDuplicateEvent
	_, errDup := engine.IngestYouTubeRevenue(ctx, rec)
	if errDup != amra.ErrDuplicateEvent {
		t.Errorf("expected ErrDuplicateEvent on duplicate ingestion, got %v", errDup)
	}

	// Test MetricsHandler
	req := httptest.NewRequest(http.MethodGet, "/api/v1/amra/metrics", nil)
	rr := httptest.NewRecorder()
	handler.MetricsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from MetricsHandler, got %d", rr.Code)
	}

	var metrics amra.FinancialMetrics
	if err := json.Unmarshal(rr.Body.Bytes(), &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	if metrics.YouTubeAccrued30d != 48.50 {
		t.Errorf("expected YouTubeAccrued30d 48.50, got %f", metrics.YouTubeAccrued30d)
	}
	if metrics.LedgerTotalTx != 1 {
		t.Errorf("expected 1 ledger transaction, got %d", metrics.LedgerTotalTx)
	}
}

func TestAMRA_GeometryHandler(t *testing.T) {
	_, handler, cleanup := setupTestAmraEngine(t)
	defer cleanup()

	// 1. GET with defaults
	req := httptest.NewRequest(http.MethodGet, "/api/v1/amra/geometry", nil)
	rr := httptest.NewRecorder()
	handler.GeometryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from GeometryHandler, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp amra.AmraGeometryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal geometry response: %v", err)
	}

	if len(resp.Body.Points) == 0 || resp.Body.SVGPath == "" {
		t.Errorf("expected non-empty body geometry: %+v", resp.Body)
	}
	if len(resp.Leaves) != 5 {
		t.Errorf("expected 5 sacred leaves, got %d", len(resp.Leaves))
	}
	if len(resp.Transmutation.Demons) != 6 {
		t.Errorf("expected 6 Arishadvarga demons, got %d", len(resp.Transmutation.Demons))
	}
	for _, d := range resp.Transmutation.Demons {
		if d.Teaching == "" {
			t.Errorf("demon %s has empty teaching", d.Name)
		}
		if d.TransmutedAs == "" {
			t.Errorf("demon %s has empty transmuted virtue", d.Name)
		}
	}
	if resp.Philosophy == nil || resp.Philosophy.KarmaPhala == "" {
		t.Errorf("expected non-empty philosophy served from BoltDB: %+v", resp.Philosophy)
	}
	if resp.Triad == nil || len(resp.Triad.Stages) != 3 {
		t.Errorf("expected 3-stage triad served from BoltDB: %+v", resp.Triad)
	}

	// 2. GET with custom parameters
	reqCustom := httptest.NewRequest(http.MethodGet, "/api/v1/amra/geometry?rx=140&hook=45&shadow=0.8&heat=0.3", nil)
	rrCustom := httptest.NewRecorder()
	handler.GeometryHandler(rrCustom, reqCustom)

	if rrCustom.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from custom GeometryHandler, got %d", rrCustom.Code)
	}

	var respCustom amra.AmraGeometryResponse
	if err := json.Unmarshal(rrCustom.Body.Bytes(), &respCustom); err != nil {
		t.Fatalf("failed to unmarshal custom geometry: %v", err)
	}
	if respCustom.Transmutation.State != "ama" {
		t.Errorf("expected state 'ama' under low heat, got '%s'", respCustom.Transmutation.State)
	}
}


