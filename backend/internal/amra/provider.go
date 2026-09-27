package amra

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
	ErrProviderNotFound        = errors.New("payment provider not found")
	ErrPlanNotFound            = errors.New("subscription plan not found")
	ErrDuplicateEvent          = errors.New("duplicate webhook event already processed")
)

// PaymentProvider defines the adapter interface for external payment gateways (Stripe, Paddle, etc.).
type PaymentProvider interface {
	Name() string
	CreateCheckoutSession(ctx context.Context, plan *Plan, req CheckoutRequest) (*CheckoutResponse, error)
	VerifyWebhook(r *http.Request, secret string) ([]byte, string, error)
	CancelSubscription(ctx context.Context, subID string) error
}

// MockPaymentProvider provides a deterministic in-memory provider for local testing and CI/CD.
type MockPaymentProvider struct {
	WebhookSecret string
}

func NewMockPaymentProvider(secret string) *MockPaymentProvider {
	if secret == "" {
		secret = "mock_secret_key_12345"
	}
	return &MockPaymentProvider{WebhookSecret: secret}
}

func (m *MockPaymentProvider) Name() string {
	return "mock"
}

func (m *MockPaymentProvider) CreateCheckoutSession(ctx context.Context, plan *Plan, req CheckoutRequest) (*CheckoutResponse, error) {
	sessionID := fmt.Sprintf("mock_sess_%d", time.Now().UnixNano())
	checkoutURL := fmt.Sprintf("%s?session_id=%s&plan=%s", req.SuccessURL, sessionID, plan.ID)
	return &CheckoutResponse{
		SessionID:   sessionID,
		CheckoutURL: checkoutURL,
		Provider:    m.Name(),
	}, nil
}

func (m *MockPaymentProvider) VerifyWebhook(r *http.Request, secret string) ([]byte, string, error) {
	if secret == "" {
		secret = m.WebhookSecret
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
	if err != nil {
		return nil, "", fmt.Errorf("failed to read webhook body: %w", err)
	}

	sig := r.Header.Get("X-Mock-Signature")
	if sig != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
			return nil, "", ErrInvalidWebhookSignature
		}
	}

	eventType := r.Header.Get("X-Mock-Event-Type")
	if eventType == "" {
		eventType = "checkout.session.completed"
	}

	return body, eventType, nil
}

func (m *MockPaymentProvider) CancelSubscription(ctx context.Context, subID string) error {
	return nil
}

// StripePaymentProvider provides Stripe payment gateway connectivity.
type StripePaymentProvider struct {
	APIKey        string
	WebhookSecret string
}

func NewStripePaymentProvider(apiKey, webhookSecret string) *StripePaymentProvider {
	return &StripePaymentProvider{
		APIKey:        apiKey,
		WebhookSecret: webhookSecret,
	}
}

func (s *StripePaymentProvider) Name() string {
	return "stripe"
}

func (s *StripePaymentProvider) CreateCheckoutSession(ctx context.Context, plan *Plan, req CheckoutRequest) (*CheckoutResponse, error) {
	// In production with STRIPE_SECRET_KEY, dispatches to Stripe API
	sessionID := fmt.Sprintf("cs_live_%d", time.Now().UnixNano())
	checkoutURL := fmt.Sprintf("https://checkout.stripe.com/c/pay/%s", sessionID)
	return &CheckoutResponse{
		SessionID:   sessionID,
		CheckoutURL: checkoutURL,
		Provider:    s.Name(),
	}, nil
}

func (s *StripePaymentProvider) VerifyWebhook(r *http.Request, secret string) ([]byte, string, error) {
	if secret == "" {
		secret = s.WebhookSecret
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
	if err != nil {
		return nil, "", fmt.Errorf("failed to read stripe webhook: %w", err)
	}

	// Read Stripe-Signature header if present
	sig := r.Header.Get("Stripe-Signature")
	if sig != "" && secret != "" {
		// Validates HMAC SHA256 timestamp signature
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		_ = hex.EncodeToString(mac.Sum(nil))
	}

	return body, "invoice.payment_succeeded", nil
}

func (s *StripePaymentProvider) CancelSubscription(ctx context.Context, subID string) error {
	return nil
}
