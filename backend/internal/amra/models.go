package amra

import "time"

// PlanTier defines subscription entitlement levels.
type PlanTier string

const (
	TierFree       PlanTier = "free"
	TierAdept      PlanTier = "adept"
	TierMagus      PlanTier = "magus"
	TierEnterprise PlanTier = "enterprise"
)

// BillingInterval defines subscription recurring cadence.
type BillingInterval string

const (
	IntervalMonthly BillingInterval = "monthly"
	IntervalYearly  BillingInterval = "yearly"
)

// SubscriptionStatus defines state of customer subscription.
type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusTrialing SubscriptionStatus = "trialing"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusCanceled SubscriptionStatus = "canceled"
)

// Plan specifies a billing product and its entitlements.
type Plan struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Tier         PlanTier        `json:"tier"`
	PriceCents   int64           `json:"price_cents"`
	Currency     string          `json:"currency"`
	Interval     BillingInterval `json:"interval"`
	Entitlements []string        `json:"entitlements"`
	MaxSeats     int             `json:"max_seats"`
}

// Subscription tracks customer entitlement and billing state.
type Subscription struct {
	ID                 string             `json:"id"`
	CustomerID         string             `json:"customer_id"`
	CustomerEmail      string             `json:"customer_email"`
	PlanID             string             `json:"plan_id"`
	Status             SubscriptionStatus `json:"status"`
	Provider           string             `json:"provider"`
	ProviderSubID      string             `json:"provider_sub_id"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CancelAtPeriodEnd  bool               `json:"cancel_at_period_end"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// Transaction represents an immutable entry in the financial audit ledger.
type Transaction struct {
	ID             string    `json:"id"`
	CustomerID     string    `json:"customer_id"`
	SubscriptionID string    `json:"subscription_id,omitempty"`
	AmountCents    int64     `json:"amount_cents"`
	Currency       string    `json:"currency"`
	Provider       string    `json:"provider"`
	Status         string    `json:"status"` // succeeded, failed, refunded, pending
	Description    string    `json:"description"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

// CheckoutRequest contains parameters for creating a payment session.
type CheckoutRequest struct {
	PlanID        string `json:"plan_id"`
	CustomerID    string `json:"customer_id"`
	CustomerEmail string `json:"customer_email"`
	SuccessURL    string `json:"success_url"`
	CancelURL     string `json:"cancel_url"`
	Provider      string `json:"provider,omitempty"` // defaults to primary adapter
}

// CheckoutResponse contains the checkout session details.
type CheckoutResponse struct {
	SessionID   string `json:"session_id"`
	CheckoutURL string `json:"checkout_url"`
	Provider    string `json:"provider"`
}
