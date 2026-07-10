package paymentgateway

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestVerifyPayinCallback(t *testing.T) {
	verifier := &WebhookVerifier{tolerance: DefaultWebhookTolerance}
	now := time.UnixMilli(1783656264000)
	values := map[string]string{
		"tradeNo":  "pay_202607101203461504780",
		"orderNo":  "PAYIN_DIRECT_1783656225381709000_38c8ac7716bd17dd",
		"currency": "USD",
		"amount":   "12.34",
		"status":   "1",
		"code":     "requires_action",
		"message":  "Processing",
	}
	query := url.Values{}
	for key, value := range values {
		query.Set(key, value)
	}
	headers := signedHeaders("1783656264000", PayinCallbackSignature("1783656264000", values))

	payload, err := verifier.VerifyPayinCallback(headers, query, now)
	if err != nil {
		t.Fatal(err)
	}
	if payload["tradeNo"] != values["tradeNo"] || payload["orderNo"] != values["orderNo"] {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestVerifyPayoutCallback(t *testing.T) {
	verifier := &WebhookVerifier{tolerance: DefaultWebhookTolerance}
	now := time.UnixMilli(1783656264000)
	values := map[string]string{
		"tradeNo":  "payout_202607101146228568143",
		"currency": "USD",
		"amount":   "3.11",
		"status":   "2",
		"code":     "succeeded",
		"message":  "TESTING: No real money will be transferred!",
	}
	query := url.Values{"orderNo": []string{"PAYOUT_1783655182791546000_61f980b238006a33"}}
	for key, value := range values {
		query.Set(key, value)
	}
	headers := signedHeaders("1783656264000", PayoutCallbackSignature("1783656264000", values))

	payload, err := verifier.VerifyPayoutCallback(headers, query, now)
	if err != nil {
		t.Fatal(err)
	}
	if payload["tradeNo"] != values["tradeNo"] || payload["orderNo"] == "" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestVerifyCallbackRejectsInvalidSignature(t *testing.T) {
	verifier := &WebhookVerifier{tolerance: DefaultWebhookTolerance}
	query := url.Values{
		"tradeNo":  []string{"pay_1"},
		"orderNo":  []string{"PAYIN_1"},
		"currency": []string{"USD"},
		"amount":   []string{"12.34"},
		"status":   []string{"1"},
		"code":     []string{"requires_action"},
		"message":  []string{"Processing"},
	}
	headers := signedHeaders("1783656264000", "bad-signature")

	if _, err := verifier.VerifyPayinCallback(headers, query, time.UnixMilli(1783656264000)); err == nil {
		t.Fatal("invalid signature should be rejected")
	}
}

func TestVerifyCallbackRejectsExpiredTimestamp(t *testing.T) {
	verifier := &WebhookVerifier{tolerance: time.Minute}
	values := map[string]string{
		"tradeNo":  "pay_1",
		"orderNo":  "PAYIN_1",
		"currency": "USD",
		"amount":   "12.34",
		"status":   "1",
		"code":     "requires_action",
		"message":  "Processing",
	}
	query := url.Values{}
	for key, value := range values {
		query.Set(key, value)
	}
	headers := signedHeaders("1783656264000", PayinCallbackSignature("1783656264000", values))

	if _, err := verifier.VerifyPayinCallback(headers, query, time.UnixMilli(1783656264000).Add(2*time.Minute)); err == nil {
		t.Fatal("expired callback timestamp should be rejected")
	}
}

func TestDebugPayoutCallbackSignature(t *testing.T) {
	verifier := &WebhookVerifier{tolerance: 0}
	values := map[string]string{
		"tradeNo":  "payout_202607101146228568143",
		"currency": "USD",
		"amount":   "3.11",
		"status":   "2",
		"code":     "succeeded",
		"message":  "TESTING: No real money will be transferred!",
	}
	query := url.Values{"orderNo": []string{"PAYOUT_1783655182791546000_61f980b238006a33"}}
	for key, value := range values {
		query.Set(key, value)
	}
	timestamp := strconvFormatUnixMilli(1783656264000)
	request := &http.Request{
		Header: signedHeaders(timestamp, PayoutCallbackSignature(timestamp, values)),
		URL:    &url.URL{RawQuery: query.Encode()},
	}

	debug := verifier.DebugPayoutCallbackSignature(request)
	if !debug.SignMatched || !debug.TimestampValid {
		t.Fatalf("signature debug should pass: %#v", debug)
	}
	if debug.SigningString == "" || debug.ExpectedSign == "" || debug.QueryValues["orderNo"] == "" {
		t.Fatalf("signature debug missing fields: %#v", debug)
	}
}

func signedHeaders(timestamp string, signature string) http.Header {
	headers := http.Header{}
	headers.Set(WebhookHeaderTimestamp, timestamp)
	headers.Set(WebhookHeaderSignature, signature)
	return headers
}

func strconvFormatUnixMilli(value int64) string {
	return strconv.FormatInt(value, 10)
}
