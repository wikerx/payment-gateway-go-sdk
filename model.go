package paymentgateway

type EncryptedRequest struct {
	Livemode bool   `json:"livemode"`
	Data     string `json:"data"`
}

type EncryptedResponse struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Livemode *bool  `json:"livemode,omitempty"`
	Data     string `json:"data,omitempty"`
}

type Result struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Livemode  *bool  `json:"livemode,omitempty"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

type APIRequest map[string]any

const (
	PaymentTypeCheckout   = "0"
	PaymentTypeDirect     = "1"
	PaymentMethodCard     = "CARD"
	PaymentMethodCashApp  = "CASHAPP"
	PaymentMethodPaypal   = "PAY_PAL"
	PaymentMethodACHDebit = "ACH_DEBIT"
	PaymentMethodUPI      = "UPI"
)

var PaymentMethods = []string{
	PaymentMethodCard,
	PaymentMethodCashApp,
	PaymentMethodPaypal,
	PaymentMethodACHDebit,
	PaymentMethodUPI,
}
