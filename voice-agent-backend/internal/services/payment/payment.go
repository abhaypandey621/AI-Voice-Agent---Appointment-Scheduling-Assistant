package payment

import (
	"fmt"
	"log"
	"time"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/charge"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/invoice"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/voice-agent/backend/internal/config"
)

// PaymentService handles payment operations via Stripe
type PaymentService struct {
	apiKey string
}

// PaymentRecord represents a payment transaction
type PaymentRecord struct {
	ID               string
	UserPhone        string
	Amount           int64 // in cents
	Currency         string
	PaymentMethod    string
	Status           string
	AppointmentID    *string
	StripeChargeID   string
	StripeCustomerID string
	StripeInvoiceID  string
	TransactionDate  time.Time
	Description      string
	Metadata         map[string]string
}

// PaymentIntent represents an intent to process payment
type PaymentIntent struct {
	ID             string
	ClientSecret   string
	Amount         int64
	Currency       string
	Status         string
	PublishableKey string
}

// NewPaymentService creates a new payment service
func NewPaymentService(cfg *config.Config) *PaymentService {
	stripe.Key = cfg.StripeSecretKey
	return &PaymentService{
		apiKey: cfg.StripeSecretKey,
	}
}

// CreatePaymentIntent creates a payment intent for appointment booking
func (ps *PaymentService) CreatePaymentIntent(userPhone, userName string, appointmentID string, amountCents int64, description string) (*PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amountCents),
		Currency:    stripe.String("usd"),
		Description: stripe.String(description),
	}

	// Add metadata
	params.AddMetadata("user_phone", userPhone)
	params.AddMetadata("user_name", userName)
	params.AddMetadata("appointment_id", appointmentID)

	pi, err := paymentintent.New(params)
	if err != nil {
		log.Printf("Failed to create payment intent: %v", err)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return &PaymentIntent{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Status:       string(pi.Status),
	}, nil
}

// ConfirmPaymentIntent confirms a payment intent
func (ps *PaymentService) ConfirmPaymentIntent(paymentIntentID string, paymentMethodID string) (*PaymentIntent, error) {
	params := &stripe.PaymentIntentConfirmParams{
		PaymentMethod: stripe.String(paymentMethodID),
	}

	pi, err := paymentintent.Confirm(paymentIntentID, params)
	if err != nil {
		log.Printf("Failed to confirm payment intent: %v", err)
		return nil, fmt.Errorf("failed to confirm payment intent: %w", err)
	}

	return &PaymentIntent{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Status:       string(pi.Status),
	}, nil
}

// ProcessPayment processes a one-time charge
func (ps *PaymentService) ProcessPayment(userPhone, userName string, amountCents int64, tokenID string, description string) (*PaymentRecord, error) {
	chargeParams := &stripe.ChargeParams{
		Amount:      stripe.Int64(amountCents),
		Currency:    stripe.String("usd"),
		Source:      &stripe.SourceParams{Token: stripe.String(tokenID)},
		Description: stripe.String(description),
	}

	chargeParams.AddMetadata("user_phone", userPhone)
	chargeParams.AddMetadata("user_name", userName)

	ch, err := charge.New(chargeParams)
	if err != nil {
		log.Printf("Failed to process charge: %v", err)
		return nil, fmt.Errorf("failed to process charge: %w", err)
	}

	return &PaymentRecord{
		ID:              ch.ID,
		UserPhone:       userPhone,
		Amount:          ch.Amount,
		Currency:        string(ch.Currency),
		Status:          string(ch.Status),
		StripeChargeID:  ch.ID,
		TransactionDate: time.Now(),
		Description:     description,
	}, nil
}

// CreateOrGetCustomer creates or retrieves a customer
func (ps *PaymentService) CreateOrGetCustomer(userPhone, userName, userEmail string) (string, error) {
	// For now, we'll create a new customer each time
	// In production, you'd want to store the customer ID
	params := &stripe.CustomerParams{
		Name:  stripe.String(userName),
		Phone: stripe.String(userPhone),
		Email: stripe.String(userEmail),
	}

	cust, err := customer.New(params)
	if err != nil {
		log.Printf("Failed to create customer: %v", err)
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	return cust.ID, nil
}

// CreateSubscription creates a recurring subscription for a customer
func (ps *PaymentService) CreateSubscription(customerID, priceID string, metadata map[string]string) (string, error) {
	// This is a placeholder for subscription creation
	// Actual implementation would use Stripe's subscription API
	log.Printf("Creating subscription for customer %s with price %s", customerID, priceID)
	return "sub_placeholder", nil
}

// CreateInvoice creates an invoice for a customer
func (ps *PaymentService) CreateInvoice(customerID, description string, items []map[string]interface{}) (string, error) {
	invoiceParams := &stripe.InvoiceParams{
		Customer:    stripe.String(customerID),
		Description: stripe.String(description),
	}

	inv, err := invoice.New(invoiceParams)
	if err != nil {
		log.Printf("Failed to create invoice: %v", err)
		return "", fmt.Errorf("failed to create invoice: %w", err)
	}

	return inv.ID, nil
}

// RefundCharge refunds a payment
func (ps *PaymentService) RefundCharge(chargeID string, reason string) error {
	// Stripe refund implementation
	log.Printf("Refunding charge %s (reason: %s)", chargeID, reason)
	// In a real implementation, use Stripe's refund API
	return nil
}

// GetPaymentStatus retrieves payment status
func (ps *PaymentService) GetPaymentStatus(chargeID string) (string, error) {
	ch, err := charge.Get(chargeID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get charge: %w", err)
	}

	return string(ch.Status), nil
}

// ValidatePaymentMethod validates a payment method
func (ps *PaymentService) ValidatePaymentMethod(token string) (bool, error) {
	// In a real implementation, you would validate the token with Stripe
	// This is just a placeholder that always returns true
	if token == "" {
		return false, fmt.Errorf("token cannot be empty")
	}
	return true, nil
}

// CalculateAppointmentCost calculates cost for an appointment
func CalculateAppointmentCost(appointmentType string, durationMinutes int) int64 {
	// Base cost in cents ($)
	baseCost := int64(1500) // $15.00

	// Add cost based on duration
	durationCost := int64(durationMinutes) * 10 // 10 cents per minute

	// Add cost based on type
	var typeMultiplier int64 = 1
	switch appointmentType {
	case "consultation":
		typeMultiplier = 1 // normal cost
	case "premium":
		typeMultiplier = 2 // double cost
	case "VIP":
		typeMultiplier = 3 // triple cost
	}

	totalCost := (baseCost + durationCost) * typeMultiplier
	return totalCost
}

// GetPayableAmount returns the amount to be paid (useful for discounts, taxes, etc.)
func GetPayableAmount(baseCost int64, discountPercent float32, taxPercent float32) int64 {
	// Apply discount
	discount := float32(baseCost) * (discountPercent / 100.0)
	afterDiscount := float32(baseCost) - discount

	// Apply tax
	tax := afterDiscount * (taxPercent / 100.0)
	total := int64(afterDiscount + tax)

	return total
}
