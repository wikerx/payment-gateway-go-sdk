package paymentgateway

import "testing"

func TestSanitizeMasksSensitiveFields(t *testing.T) {
	value := Sanitize(map[string]any{
		"number": "4000056655665556",
		"cvc":    "123",
		"nested": map[string]any{"email": "lily@example.com"},
	}).(map[string]any)
	if value["number"] == "4000056655665556" || value["cvc"] == "123" {
		t.Fatal("sensitive fields were not masked")
	}
}
