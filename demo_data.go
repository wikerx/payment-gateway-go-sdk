package paymentgateway

// PaymentMethodDataExamples returns sandbox-only paymentMethodData examples
// used by the demo console and sample programs. Merchants should replace these
// values with their own customer/payment details in real integrations.
func PaymentMethodDataExamples() map[string]map[string]any {
	return map[string]map[string]any{
		PaymentMethodCashApp: {
			"cashappAccount": "$123",
			"email":          "lily_brown_1782457030419@test.com",
		},
		PaymentMethodCard: {
			"number":     "4000056655665556",
			"expMonth":   "06",
			"expYear":    "2029",
			"cvc":        "123",
			"email":      "lily_brown_1782457030419@test.com",
			"holderName": "Lily Brown",
		},
		PaymentMethodPaypal: {
			"paypalEmail": "lily_brown_1782457030419@test.com",
		},
		PaymentMethodACHDebit: {
			"accountNumber": "6205500000000000004",
			"routingNumber": "641110",
		},
		PaymentMethodUPI: {
			"bankName": "scott's bank",
			"bankCode": "641110",
			"cardNo":   "6200000000000005",
		},
	}
}

// CustomerExample returns a sandbox customer object used by sample requests.
func CustomerExample() map[string]any {
	return map[string]any{
		"firstname": "Lily",
		"lastname":  "Brown",
		"email":     "lily_brown_1782457030419@test.com",
		"phone":     "13628173752",
		"country":   "US",
		"state":     "CA",
		"city":      "Los Angeles",
		"address":   "123 Main St, Apt 4B",
		"zipcode":   "90001",
	}
}
