package amra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// YouTubeRevenueRecord represents verified daily revenue from YouTube to be ledged into AMRA.
type YouTubeRevenueRecord struct {
	Day              string  `json:"day"` // YYYY-MM-DD
	EstimatedRevenue float64 `json:"estimated_revenue"`
	MonetizedPlays   int64   `json:"monetized_playbacks"`
	CPM              float64 `json:"cpm"`
	ChannelID        string  `json:"channel_id"`
}

// IngestYouTubeRevenue records verified YouTube revenue into the immutable AMRA financial ledger.
func (e *Engine) IngestYouTubeRevenue(_ context.Context, rec YouTubeRevenueRecord) (*Transaction, error) {
	if rec.Day == "" {
		return nil, fmt.Errorf("day is required")
	}

	idempotencyKey := fmt.Sprintf("idem_yt_rev_%s_%s", rec.ChannelID, rec.Day)

	// Deduplicate via AMRA idempotency bucket
	if _, err := e.store.GetJSON(db.BucketAmraIdempotency, idempotencyKey); err == nil {
		return nil, ErrDuplicateEvent
	}

	amountCents := int64(rec.EstimatedRevenue * 100)
	now := time.Now().UTC()
	txID := fmt.Sprintf("tx_yt_%s_%d", rec.Day, now.UnixNano())

	tx := &Transaction{
		ID:             txID,
		CustomerID:     fmt.Sprintf("channel:%s", rec.ChannelID),
		AmountCents:    amountCents,
		Currency:       "USD",
		Provider:       "youtube_partner",
		Status:         "accrued",
		Description:    fmt.Sprintf("YouTube Partner Ad Revenue for %s (Plays: %d, CPM: $%.2f)", rec.Day, rec.MonetizedPlays, rec.CPM),
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	if err := e.store.PutJSON(db.BucketAmraLedger, tx.ID, txBytes); err != nil {
		return nil, fmt.Errorf("failed to put transaction in ledger: %w", err)
	}

	idemRecord := map[string]any{
		"idempotency_key": idempotencyKey,
		"processed_at":    now,
		"tx_id":           tx.ID,
		"amount_cents":    amountCents,
	}
	idemBytes, _ := json.Marshal(idemRecord)
	_ = e.store.PutJSON(db.BucketAmraIdempotency, idempotencyKey, idemBytes)

	return tx, nil
}

// FinancialMetrics consolidates SaaS subscriptions and YouTube partner revenue.
type FinancialMetrics struct {
	ActiveSubscribers    int     `json:"active_subscribers"`
	SaaSMonthlyRunRate   float64 `json:"saas_mrr"`
	YouTubeAccrued30d    float64 `json:"youtube_accrued_30d"`
	TotalGrossEcosystem  float64 `json:"total_gross_ecosystem"`
	LedgerTotalTx        int     `json:"ledger_total_transactions"`
	Timestamp            string  `json:"timestamp"`
}

// GetFinancialMetrics calculates unified ecosystem financial telemetry.
func (e *Engine) GetFinancialMetrics(_ context.Context) (*FinancialMetrics, error) {
	// 1. Calculate active SaaS subscribers & MRR
	subKeys, err := e.store.ListKeys(db.BucketAmraSubscriptions, "sub_")
	activeCount := 0
	var saasMRR int64

	if err == nil {
		for _, k := range subKeys {
			data, err := e.store.GetJSON(db.BucketAmraSubscriptions, k)
			if err != nil {
				continue
			}
			var sub Subscription
			if err := json.Unmarshal(data, &sub); err == nil && sub.Status == StatusActive {
				activeCount++
				if p, ok := e.plans[sub.PlanID]; ok {
					if p.Interval == IntervalYearly {
						saasMRR += p.PriceCents / 12
					} else {
						saasMRR += p.PriceCents
					}
				}
			}
		}
	}

	// 2. Calculate YouTube revenue from ledger
	txKeys, err := e.store.ListKeys(db.BucketAmraLedger, "tx_")
	var ytAccrued int64
	totalTx := len(txKeys)

	if err == nil {
		for _, k := range txKeys {
			data, err := e.store.GetJSON(db.BucketAmraLedger, k)
			if err != nil {
				continue
			}
			var tx Transaction
			if err := json.Unmarshal(data, &tx); err == nil && tx.Provider == "youtube_partner" {
				ytAccrued += tx.AmountCents
			}
		}
	}

	saasDollars := float64(saasMRR) / 100.0
	ytDollars := float64(ytAccrued) / 100.0

	return &FinancialMetrics{
		ActiveSubscribers:    activeCount,
		SaaSMonthlyRunRate:   saasDollars,
		YouTubeAccrued30d:    ytDollars,
		TotalGrossEcosystem:  saasDollars + ytDollars,
		LedgerTotalTx:        totalTx,
		Timestamp:            time.Now().UTC().Format(time.RFC3339),
	}, nil
}
