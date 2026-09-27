package amra_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/amra"
)

func TestAMRA_GCloudStatus(t *testing.T) {
	_, handler, cleanup := setupTestAmraEngine(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/amra/gcloud/status", nil)
	rec := httptest.NewRecorder()

	handler.GCloudStatusHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from GCloudStatusHandler, got %d", rec.Code)
	}

	var resp amra.GCloudStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse GCloud status response: %v", err)
	}

	if resp.Account.Account == "" {
		t.Error("expected non-empty account identity")
	}
	if resp.Project.ProjectID == "" {
		t.Error("expected non-empty project ID")
	}
	if !resp.BillingAccount.Open {
		t.Error("expected billing account to be open")
	}
}

func TestAMRA_GCloudBillingAndLedgerSync(t *testing.T) {
	engine, handler, cleanup := setupTestAmraEngine(t)
	defer cleanup()

	// 1. Test Billing Report
	req := httptest.NewRequest(http.MethodGet, "/api/v1/amra/gcloud/billing", nil)
	rec := httptest.NewRecorder()

	handler.GCloudBillingHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from GCloudBillingHandler, got %d", rec.Code)
	}

	var billing amra.GCloudBillingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &respParseWrapper{Data: &billing}); err != nil {
		if err2 := json.Unmarshal(rec.Body.Bytes(), &billing); err2 != nil {
			t.Fatalf("failed to parse billing response: %v", err2)
		}
	}

	if billing.CostReport.AccruedCostDollars <= 0 {
		t.Errorf("expected accrued cost > 0, got %f", billing.CostReport.AccruedCostDollars)
	}
	if len(billing.CostReport.Services) == 0 {
		t.Error("expected itemized services in billing report")
	}
	if billing.CostReport.BudgetLimitDollars != 100.0 {
		t.Errorf("expected budget limit 100.00, got %f", billing.CostReport.BudgetLimitDollars)
	}
	if !billing.CostReport.ScaleToZeroOptimized {
		t.Error("expected scale-to-zero optimization to be true")
	}

	// 2. Test Ledger Expense Sync
	syncReq := httptest.NewRequest(http.MethodPost, "/api/v1/amra/gcloud/sync-ledger?month=2026-09", nil)
	syncRec := httptest.NewRecorder()

	handler.GCloudSyncLedgerHandler(syncRec, syncReq)

	if syncRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from GCloudSyncLedgerHandler, got %d: %s", syncRec.Code, syncRec.Body.String())
	}

	var syncResp struct {
		Status      string            `json:"status"`
		Transaction *amra.Transaction `json:"transaction"`
	}
	if err := json.Unmarshal(syncRec.Body.Bytes(), &syncResp); err != nil {
		t.Fatalf("failed to parse sync response: %v", err)
	}

	if syncResp.Status != "synced" {
		t.Errorf("expected status 'synced', got '%s'", syncResp.Status)
	}
	if syncResp.Transaction == nil {
		t.Fatal("expected non-nil transaction in sync response")
	}
	if syncResp.Transaction.Provider != "gcloud" {
		t.Errorf("expected provider 'gcloud', got '%s'", syncResp.Transaction.Provider)
	}
	if syncResp.Transaction.AmountCents >= 0 {
		t.Errorf("expected negative amount cents for expense, got %d", syncResp.Transaction.AmountCents)
	}

	// 3. Verify Idempotent Deduplication (syncing again returns existing transaction)
	syncReq2 := httptest.NewRequest(http.MethodPost, "/api/v1/amra/gcloud/sync-ledger?month=2026-09", nil)
	syncRec2 := httptest.NewRecorder()

	handler.GCloudSyncLedgerHandler(syncRec2, syncReq2)
	if syncRec2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from duplicate sync, got %d", syncRec2.Code)
	}

	// 4. Verify Ledger Retrieval
	txs, err := engine.ListLedgerTransactions(10)
	if err != nil {
		t.Fatalf("failed to list ledger transactions: %v", err)
	}
	found := false
	for _, tx := range txs {
		if tx.Provider == "gcloud" && tx.ID == syncResp.Transaction.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected gcloud expense transaction to appear in ledger list")
	}

	// 5. Verify Billing Report now indicates ledger_synced = true
	billing2, err := engine.GetGCloudBilling(context.Background())
	if err != nil {
		t.Fatalf("failed to re-fetch billing: %v", err)
	}
	if !billing2.LedgerSynced {
		t.Error("expected LedgerSynced to be true after sync")
	}
}

type respParseWrapper struct {
	Data any
}

func (w *respParseWrapper) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, w.Data)
}
