package amra

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// Engine orchestrates subscriptions, providers, webhook idempotency, and the financial audit ledger.
type Engine struct {
	store     db.StorageEngine
	providers map[string]PaymentProvider
	plans     map[string]Plan
}

// NewEngine initializes the AMRA financial core with default plan catalogs and registered payment providers.
func NewEngine(store db.StorageEngine, providers ...PaymentProvider) *Engine {
	e := &Engine{
		store:     store,
		providers: make(map[string]PaymentProvider),
		plans:     make(map[string]Plan),
	}

	// Register providers
	for _, p := range providers {
		e.providers[p.Name()] = p
	}
	if len(e.providers) == 0 {
		mock := NewMockPaymentProvider("")
		e.providers[mock.Name()] = mock
	}

	// Seed canonical plan catalog
	e.seedPlans()

	return e
}

func (e *Engine) seedPlans() {
	defaultPlans := []Plan{
		{
			ID:         "plan_adept_monthly",
			Name:       "Adept (Monthly)",
			Tier:       TierAdept,
			PriceCents: 2900,
			Currency:   "USD",
			Interval:   IntervalMonthly,
			Entitlements: []string{
				"ephemeris:realtime",
				"nakshatras:full_calculations",
				"dasha:mahadasha_timelines",
			},
			MaxSeats: 1,
		},
		{
			ID:         "plan_adept_yearly",
			Name:       "Adept (Annual)",
			Tier:       TierAdept,
			PriceCents: 29000,
			Currency:   "USD",
			Interval:   IntervalYearly,
			Entitlements: []string{
				"ephemeris:realtime",
				"nakshatras:full_calculations",
				"dasha:mahadasha_timelines",
			},
			MaxSeats: 1,
		},
		{
			ID:         "plan_magus_monthly",
			Name:       "Magus (Monthly)",
			Tier:       TierMagus,
			PriceCents: 8900,
			Currency:   "USD",
			Interval:   IntervalMonthly,
			Entitlements: []string{
				"ephemeris:realtime",
				"nakshatras:full_calculations",
				"dasha:mahadasha_timelines",
				"alchemy:hermetic_models",
				"axis_mundi:voice_triage",
				"boltyaml:yaml_ingestion",
			},
			MaxSeats: 5,
		},
		{
			ID:         "plan_enterprise",
			Name:       "Enterprise Sovereign",
			Tier:       TierEnterprise,
			PriceCents: 29900,
			Currency:   "USD",
			Interval:   IntervalMonthly,
			Entitlements: []string{
				"all_features",
				"cloudrun:dedicated_instance",
				"sla:99_99",
				"audit_ledger:realtime_streaming",
			},
			MaxSeats: 50,
		},
	}

	for _, p := range defaultPlans {
		e.plans[p.ID] = p
	}
}

// ListPlans returns all available billing plans.
func (e *Engine) ListPlans() []Plan {
	res := make([]Plan, 0, len(e.plans))
	for _, p := range e.plans {
		res = append(res, p)
	}
	return res
}

// GetPlan retrieves a specific plan by ID.
func (e *Engine) GetPlan(id string) (*Plan, error) {
	p, ok := e.plans[id]
	if !ok {
		return nil, ErrPlanNotFound
	}
	return &p, nil
}

// CreateCheckout initiates a payment session with the chosen or default provider.
func (e *Engine) CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error) {
	plan, err := e.GetPlan(req.PlanID)
	if err != nil {
		return nil, err
	}

	providerName := req.Provider
	if providerName == "" {
		// Default to mock or first registered provider
		if _, ok := e.providers["stripe"]; ok {
			providerName = "stripe"
		} else {
			providerName = "mock"
		}
	}

	provider, ok := e.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerName)
	}

	return provider.CreateCheckoutSession(ctx, plan, req)
}

// ProcessWebhookIngest handles incoming payment events, verifies idempotency, updates subscriptions, and writes to the ledger.
func (e *Engine) ProcessWebhookIngest(ctx context.Context, providerName string, eventType string, rawBody []byte) (*Transaction, error) {
	// 1. Calculate Idempotency Hash
	hash := sha256.Sum256(rawBody)
	idempotencyKey := fmt.Sprintf("%s:%s", providerName, hex.EncodeToString(hash[:16]))

	// Check if already processed in BoltDB
	if _, err := e.store.GetJSON(db.BucketAmraIdempotency, idempotencyKey); err == nil {
		return nil, ErrDuplicateEvent
	}

	// 2. Parse mock/generic payload
	var payload struct {
		CustomerID     string `json:"customer_id"`
		CustomerEmail  string `json:"customer_email"`
		PlanID         string `json:"plan_id"`
		AmountCents    int64  `json:"amount_cents"`
		Currency       string `json:"currency"`
		SubscriptionID string `json:"subscription_id"`
	}
	_ = json.Unmarshal(rawBody, &payload)

	if payload.CustomerID == "" {
		payload.CustomerID = "cust_anonymous"
	}
	if payload.PlanID == "" {
		payload.PlanID = "plan_adept_monthly"
	}
	if payload.Currency == "" {
		payload.Currency = "USD"
	}
	if payload.AmountCents == 0 {
		if p, err := e.GetPlan(payload.PlanID); err == nil {
			payload.AmountCents = p.PriceCents
		} else {
			payload.AmountCents = 2900
		}
	}

	subID := payload.SubscriptionID
	if subID == "" {
		subID = fmt.Sprintf("sub_%s_%d", payload.CustomerID, time.Now().Unix())
	}

	// 3. Upsert Subscription in BoltDB
	now := time.Now().UTC()
	sub := &Subscription{
		ID:                 subID,
		CustomerID:         payload.CustomerID,
		CustomerEmail:      payload.CustomerEmail,
		PlanID:             payload.PlanID,
		Status:             StatusActive,
		Provider:           providerName,
		ProviderSubID:      fmt.Sprintf("%s_%d", providerName, now.Unix()),
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	subBytes, _ := json.Marshal(sub)
	if err := e.store.PutJSON(db.BucketAmraSubscriptions, sub.ID, subBytes); err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	// 4. Record Immutable Financial Ledger Transaction
	txID := fmt.Sprintf("tx_%d", now.UnixNano())
	tx := &Transaction{
		ID:             txID,
		CustomerID:     payload.CustomerID,
		SubscriptionID: sub.ID,
		AmountCents:    payload.AmountCents,
		Currency:       payload.Currency,
		Provider:       providerName,
		Status:         "succeeded",
		Description:    fmt.Sprintf("Payment for %s via %s", payload.PlanID, providerName),
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
	}

	txBytes, _ := json.Marshal(tx)
	if err := e.store.PutJSON(db.BucketAmraLedger, tx.ID, txBytes); err != nil {
		return nil, fmt.Errorf("failed to record ledger transaction: %w", err)
	}

	// 5. Mark Idempotency Processed
	idemRecord := map[string]any{
		"idempotency_key": idempotencyKey,
		"processed_at":    now,
		"tx_id":           tx.ID,
	}
	idemBytes, _ := json.Marshal(idemRecord)
	_ = e.store.PutJSON(db.BucketAmraIdempotency, idempotencyKey, idemBytes)

	return tx, nil
}

// GetSubscription fetches subscription details by ID.
func (e *Engine) GetSubscription(id string) (*Subscription, error) {
	data, err := e.store.GetJSON(db.BucketAmraSubscriptions, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, errors.New("subscription not found")
		}
		return nil, err
	}

	var sub Subscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListLedgerTransactions retrieves the most recent transactions from the immutable ledger.
func (e *Engine) ListLedgerTransactions(limit int) ([]*Transaction, error) {
	keys, err := e.store.ListKeys(db.BucketAmraLedger, "tx_")
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(keys) {
		limit = len(keys)
	}

	res := make([]*Transaction, 0, limit)
	for i := len(keys) - 1; i >= 0 && len(res) < limit; i-- {
		data, err := e.store.GetJSON(db.BucketAmraLedger, keys[i])
		if err != nil {
			continue
		}
		var tx Transaction
		if err := json.Unmarshal(data, &tx); err == nil {
			res = append(res, &tx)
		}
	}

	return res, nil
}

// GetEsotericDoc retrieves an arbitrary esoteric document from BoltDB.
func (e *Engine) GetEsotericDoc(key string) ([]byte, error) {
	return e.store.GetEsotericContent(key)
}

// GetArishadvargaDemons loads the 6 classical inner adversaries and transmutations from BoltDB.
func (e *Engine) GetArishadvargaDemons() ([]db.ArishadvargaDemon, error) {
	raw, err := e.store.GetEsotericContent("esoteric:arishadvarga")
	if err != nil {
		return db.DefaultArishadvargaDemons(), nil
	}
	var demons []db.ArishadvargaDemon
	if err := json.Unmarshal(raw, &demons); err != nil {
		return db.DefaultArishadvargaDemons(), nil
	}
	return demons, nil
}

// GetAmraPhilosophy loads the canonical Vedic philosophy document from BoltDB.
func (e *Engine) GetAmraPhilosophy() (*db.EsotericPhilosophyDoc, error) {
	raw, err := e.store.GetEsotericContent("esoteric:amra_philosophy")
	if err != nil {
		doc := db.DefaultAmraPhilosophyDoc()
		return &doc, nil
	}
	var doc db.EsotericPhilosophyDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		def := db.DefaultAmraPhilosophyDoc()
		return &def, nil
	}
	return &doc, nil
}

// GetTransmutationTriad loads the 3-stage alchemical progression from BoltDB.
func (e *Engine) GetTransmutationTriad() (*db.TransmutationTriadDoc, error) {
	raw, err := e.store.GetEsotericContent("esoteric:transmutation_triad")
	if err != nil {
		doc := db.DefaultTransmutationTriadDoc()
		return &doc, nil
	}
	var doc db.TransmutationTriadDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		def := db.DefaultTransmutationTriadDoc()
		return &def, nil
	}
	return &doc, nil
}

