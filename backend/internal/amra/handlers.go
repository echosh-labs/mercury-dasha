package amra

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// Handler serves HTTP endpoints for AMRA financial operations.
type Handler struct {
	engine *Engine
}

func NewHandler(engine *Engine) *Handler {
	return &Handler{engine: engine}
}

// ListPlansHandler returns all active subscription plans.
func (h *Handler) ListPlansHandler(w http.ResponseWriter, r *http.Request) {
	plans := h.engine.ListPlans()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"plans": plans,
		"count": len(plans),
	})
}

// CheckoutHandler creates a checkout session for a given plan.
func (h *Handler) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.PlanID == "" {
		http.Error(w, `{"error":"plan_id is required"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.engine.CreateCheckout(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			http.Error(w, `{"error":"plan not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// WebhookHandler handles inbound payment provider webhook events.
func (h *Handler) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	if providerName == "" {
		providerName = "mock"
	}

	provider, ok := h.engine.providers[providerName]
	if !ok {
		http.Error(w, `{"error":"unrecognized payment provider"}`, http.StatusBadRequest)
		return
	}

	body, eventType, err := provider.VerifyWebhook(r, "")
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	tx, err := h.engine.ProcessWebhookIngest(r.Context(), providerName, eventType, body)
	if err != nil {
		if errors.Is(err, ErrDuplicateEvent) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ignored_duplicate"})
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "processed",
		"transaction": tx,
	})
}

// GetSubscriptionHandler retrieves subscription state by ID.
func (h *Handler) GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	subID := r.PathValue("id")
	if subID == "" {
		http.Error(w, `{"error":"subscription id is required"}`, http.StatusBadRequest)
		return
	}

	sub, err := h.engine.GetSubscription(subID)
	if err != nil {
		http.Error(w, `{"error":"subscription not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sub)
}

// LedgerHandler returns recent transactions from the immutable ledger.
func (h *Handler) LedgerHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	txs, err := h.engine.ListLedgerTransactions(limit)
	if err != nil {
		http.Error(w, `{"error":"failed to query ledger"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ledger": txs,
		"count":  len(txs),
	})
}

// MetricsHandler returns unified SaaS and YouTube financial metrics.
func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.engine.GetFinancialMetrics(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to compute financial metrics"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(metrics)
}

// GeometryHandler calculates the parametric Kairi curve, sacred leaves, and
// Samudra Manthan alchemical shadow-transmutation metrics.
func (h *Handler) GeometryHandler(w http.ResponseWriter, r *http.Request) {
	params := DefaultKairiParams()
	shadowTension := 0.40
	solarFire := 0.80

	// Handle optional POST request body
	if r.Method == http.MethodPost && r.Body != nil {
		var req struct {
			Rx            float64 `json:"rx"`
			Ry            float64 `json:"ry"`
			Hook          float64 `json:"hook"`
			ShadowTension float64 `json:"shadow_tension"`
			SolarFire     float64 `json:"solar_fire"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.Rx > 0 {
				params.Rx = req.Rx
			}
			if req.Ry > 0 {
				params.Ry = req.Ry
			}
			if req.Hook > 0 {
				params.Gamma = req.Hook
			}
			if req.ShadowTension >= 0 {
				shadowTension = req.ShadowTension
			}
			if req.SolarFire >= 0 {
				solarFire = req.SolarFire
			}
		}
	}

	// Handle optional GET query parameters
	q := r.URL.Query()
	if rxStr := q.Get("rx"); rxStr != "" {
		if parsed, err := strconv.ParseFloat(rxStr, 64); err == nil && parsed > 0 {
			params.Rx = parsed
		}
	}
	if ryStr := q.Get("ry"); ryStr != "" {
		if parsed, err := strconv.ParseFloat(ryStr, 64); err == nil && parsed > 0 {
			params.Ry = parsed
		}
	}
	if hookStr := q.Get("hook"); hookStr != "" {
		if parsed, err := strconv.ParseFloat(hookStr, 64); err == nil && parsed > 0 {
			params.Gamma = parsed
		}
	}
	if shadowStr := q.Get("shadow"); shadowStr != "" {
		if parsed, err := strconv.ParseFloat(shadowStr, 64); err == nil {
			shadowTension = parsed
		}
	}
	if heatStr := q.Get("heat"); heatStr != "" {
		if parsed, err := strconv.ParseFloat(heatStr, 64); err == nil {
			solarFire = parsed
		}
	}

	demons, _ := h.engine.GetArishadvargaDemons()
	philosophy, _ := h.engine.GetAmraPhilosophy()
	triad, _ := h.engine.GetTransmutationTriad()

	geo := GenerateAmraGeometryWithContent(params, shadowTension, solarFire, demons, philosophy, triad)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(geo)
}

// GCloudStatusHandler returns the Google Cloud account and project topology.
func (h *Handler) GCloudStatusHandler(w http.ResponseWriter, r *http.Request) {
	status, err := h.engine.GetGCloudStatus(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to resolve gcloud status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}

// GCloudBillingHandler returns the infrastructure burn rate report and sovereign margin analysis.
func (h *Handler) GCloudBillingHandler(w http.ResponseWriter, r *http.Request) {
	billing, err := h.engine.GetGCloudBilling(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to generate gcloud billing report"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(billing)
}

// GCloudSyncLedgerHandler commits the current month's infrastructure expense to the immutable ledger.
func (h *Handler) GCloudSyncLedgerHandler(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	tx, err := h.engine.SyncGCloudExpenseToLedger(r.Context(), month)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to sync gcloud expense to ledger: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "synced",
		"transaction": tx,
	})
}
