package payment

import "context"

// Gateway abstracts payment provider (e.g., Stripe)

type Gateway interface {
	Charge(ctx context.Context, amountCents int64, method string, metadata map[string]string) (paymentRef string, err error)
}

// MockGateway simulates a payment provider.

type MockGateway struct{}

func (m *MockGateway) Charge(ctx context.Context, amountCents int64, method string, metadata map[string]string) (string, error) {
	// Always succeed and return a mock reference
	return "pay_" + metadata["order_id"], nil
}
