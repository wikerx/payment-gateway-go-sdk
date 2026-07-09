package paymentgateway

// EncryptedRequest is the HTTP request envelope sent to OpenAPI endpoints with
// a body. Data is the compact encrypted payload.
type EncryptedRequest struct {
	Livemode bool   `json:"livemode"`
	Data     string `json:"data"`
}

// EncryptedResponse is the raw gateway response envelope before SDK data
// decryption. Data may be blank for responses without a business payload.
type EncryptedResponse struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Livemode *bool  `json:"livemode,omitempty"`
	Data     string `json:"data,omitempty"`
}

// Result is the SDK-facing response after encrypted data has been decrypted and
// parsed. Code and Msg still come from the gateway business envelope.
type Result struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Livemode  *bool  `json:"livemode,omitempty"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// APIRequest is a flexible JSON object used by examples and demo pages. It lets
// merchants pass the exact OpenAPI request fields documented for each endpoint.
type APIRequest map[string]any

const (
	// PaymentTypeCheckout creates a hosted checkout payment.
	PaymentTypeCheckout = "0"
	// PaymentTypeDirect creates a direct/local payment.
	PaymentTypeDirect = "1"
	// PaymentMethodCard identifies card payment method data.
	PaymentMethodCard = "CARD"
	// PaymentMethodCashApp identifies Cash App payment method data.
	PaymentMethodCashApp = "CASHAPP"
	// PaymentMethodPaypal identifies PayPal payment method data.
	PaymentMethodPaypal = "PAY_PAL"
	// PaymentMethodACHDebit identifies ACH debit payment method data.
	PaymentMethodACHDebit = "ACH_DEBIT"
	// PaymentMethodUPI identifies UPI/bank-card style payout method data.
	PaymentMethodUPI = "UPI"
)

// PaymentMethods lists the payment method codes supported by the demo helpers.
var PaymentMethods = []string{
	PaymentMethodCard,
	PaymentMethodCashApp,
	PaymentMethodPaypal,
	PaymentMethodACHDebit,
	PaymentMethodUPI,
}
