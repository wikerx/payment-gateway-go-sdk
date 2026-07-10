package paymentgateway

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// WebhookHeaderTimestamp is the callback header that carries the gateway
	// signing timestamp in milliseconds.
	WebhookHeaderTimestamp = "t"
	// WebhookHeaderSignature is the callback header that carries the expected
	// SHA256 hex signature.
	WebhookHeaderSignature = "signature"
	// DefaultWebhookTolerance is the default allowed clock skew for callback
	// signature verification.
	DefaultWebhookTolerance = 5 * time.Minute
)

type WebhookVerifier struct {
	payloadCrypto *PayloadCrypto
	privateKey    string
	livemode      bool
	tolerance     time.Duration
}

// WebhookSignatureDebug describes exactly how a callback signature is checked.
// It is safe to print in sandbox logs because it contains callback parameters
// and signatures, but no private keys.
type WebhookSignatureDebug struct {
	Timestamp        string            `json:"timestamp"`
	ReceivedSign     string            `json:"receivedSign"`
	ExpectedSign     string            `json:"expectedSign"`
	SigningString    string            `json:"signingString"`
	SignMatched      bool              `json:"signMatched"`
	TimestampValid   bool              `json:"timestampValid"`
	TimestampError   string            `json:"timestampError,omitempty"`
	SignedFieldNames []string          `json:"signedFieldNames"`
	SignedValues     map[string]string `json:"signedValues"`
	QueryValues      map[string]string `json:"queryValues"`
}

// NewWebhookVerifier creates a verifier for encrypted gateway webhook bodies.
// It reuses the merchant response private key from Config.
func NewWebhookVerifier(config *Config) (*WebhookVerifier, error) {
	if config == nil {
		return nil, configError("config can not be nil", nil)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &WebhookVerifier{
		payloadCrypto: NewPayloadCrypto(),
		privateKey:    config.MerchantResponsePrivateKey,
		livemode:      config.Livemode,
		tolerance:     DefaultWebhookTolerance,
	}, nil
}

// Verify parses the webhook response envelope, checks livemode when present,
// decrypts data, and returns the business payload as a JSON object.
func (v *WebhookVerifier) Verify(rawBody []byte) (map[string]any, error) {
	if len(rawBody) == 0 {
		return nil, validationError("webhook body can not be blank")
	}
	var envelope EncryptedResponse
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return nil, responseError("webhook encrypted body can not be parsed", err)
	}
	if envelope.Livemode != nil && *envelope.Livemode != v.livemode {
		return nil, responseError("webhook livemode is inconsistent", nil)
	}
	if strings.TrimSpace(envelope.Data) == "" {
		return nil, responseError("webhook encrypted data can not be blank", nil)
	}
	// Webhook data is encrypted the same way as synchronous response data: the
	// merchant response private key unwraps the AES content key.
	privateKey, err := ReadPrivateKey(v.privateKey)
	if err != nil {
		return nil, err
	}
	plainJSON, err := v.payloadCrypto.Decrypt(envelope.Data, privateKey)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := fromJSON(plainJSON, &payload); err != nil {
		return nil, responseError("webhook payload can not be parsed", err)
	}
	return payload, nil
}

// SetTolerance changes the allowed difference between the local clock and the
// callback header timestamp. A non-positive value disables timestamp freshness
// checks, which is useful only for controlled tests.
func (v *WebhookVerifier) SetTolerance(tolerance time.Duration) {
	v.tolerance = tolerance
}

// VerifyPayinCallbackRequest verifies a payin callback sent as GET query
// parameters. It checks the t/signature headers using the documented payin
// signing string:
//
//	t + tradeNo + orderNo + currency + amount + status + code + message
//
// The returned map contains the callback query parameters exactly as received
// after URL decoding. Merchants should still apply business idempotency before
// updating local order state.
func (v *WebhookVerifier) VerifyPayinCallbackRequest(request *http.Request) (map[string]string, error) {
	if request == nil {
		return nil, validationError("webhook request can not be nil")
	}
	return v.VerifyPayinCallback(request.Header, request.URL.Query(), time.Now())
}

// VerifyPayoutCallbackRequest verifies a payout callback sent as GET query
// parameters. It checks the t/signature headers using the documented payout
// signing string:
//
//	t + tradeNo + currency + amount + status + code + message
//
// The payout callback may include orderNo in the query, but orderNo is not part
// of the documented payout signature string.
func (v *WebhookVerifier) VerifyPayoutCallbackRequest(request *http.Request) (map[string]string, error) {
	if request == nil {
		return nil, validationError("webhook request can not be nil")
	}
	return v.VerifyPayoutCallback(request.Header, request.URL.Query(), time.Now())
}

// VerifyPayinCallback verifies payin callback headers and query parameters.
// It is useful when merchants use a web framework that exposes headers/query
// separately instead of passing *http.Request through.
func (v *WebhookVerifier) VerifyPayinCallback(headers http.Header, query url.Values, now time.Time) (map[string]string, error) {
	return verifySignedWebhook(headers, query, now, v.tolerance, []string{
		"tradeNo",
		"orderNo",
		"currency",
		"amount",
		"status",
		"code",
		"message",
	})
}

// VerifyPayoutCallback verifies payout callback headers and query parameters.
// It is useful when merchants use a web framework that exposes headers/query
// separately instead of passing *http.Request through.
func (v *WebhookVerifier) VerifyPayoutCallback(headers http.Header, query url.Values, now time.Time) (map[string]string, error) {
	return verifySignedWebhook(headers, query, now, v.tolerance, []string{
		"tradeNo",
		"currency",
		"amount",
		"status",
		"code",
		"message",
	})
}

// DebugPayinCallbackSignature builds a printable signature verification report
// for payin callbacks. It does not return an error for invalid signatures; the
// caller can print the report before calling VerifyPayinCallbackRequest.
func (v *WebhookVerifier) DebugPayinCallbackSignature(request *http.Request) WebhookSignatureDebug {
	if request == nil {
		return WebhookSignatureDebug{TimestampError: "webhook request can not be nil"}
	}
	return buildWebhookSignatureDebug(request.Header, request.URL.Query(), time.Now(), v.tolerance, []string{
		"tradeNo",
		"orderNo",
		"currency",
		"amount",
		"status",
		"code",
		"message",
	})
}

// DebugPayoutCallbackSignature builds a printable signature verification report
// for payout callbacks. The payout signature string intentionally excludes
// orderNo, matching the public callback document.
func (v *WebhookVerifier) DebugPayoutCallbackSignature(request *http.Request) WebhookSignatureDebug {
	if request == nil {
		return WebhookSignatureDebug{TimestampError: "webhook request can not be nil"}
	}
	return buildWebhookSignatureDebug(request.Header, request.URL.Query(), time.Now(), v.tolerance, []string{
		"tradeNo",
		"currency",
		"amount",
		"status",
		"code",
		"message",
	})
}

// PayinCallbackSignature returns the expected SHA256 hex signature for a payin
// callback. It is exported so merchants can test their own handlers without
// copying the signing algorithm.
func PayinCallbackSignature(timestamp string, values map[string]string) string {
	return callbackSignature(timestamp, values, []string{"tradeNo", "orderNo", "currency", "amount", "status", "code", "message"})
}

// PayoutCallbackSignature returns the expected SHA256 hex signature for a
// payout callback. The documented payout signature does not include orderNo.
func PayoutCallbackSignature(timestamp string, values map[string]string) string {
	return callbackSignature(timestamp, values, []string{"tradeNo", "currency", "amount", "status", "code", "message"})
}

func verifySignedWebhook(headers http.Header, query url.Values, now time.Time, tolerance time.Duration, signedFields []string) (map[string]string, error) {
	timestamp := strings.TrimSpace(headers.Get(WebhookHeaderTimestamp))
	if timestamp == "" {
		return nil, validationError("webhook header t can not be blank")
	}
	signature := strings.TrimSpace(headers.Get(WebhookHeaderSignature))
	if signature == "" {
		return nil, validationError("webhook header signature can not be blank")
	}
	if tolerance > 0 {
		if err := validateWebhookTimestamp(timestamp, now, tolerance); err != nil {
			return nil, err
		}
	}
	values, err := requiredQueryValues(query, signedFields)
	if err != nil {
		return nil, err
	}
	expected := callbackSignature(timestamp, values, signedFields)
	if subtle.ConstantTimeCompare([]byte(strings.ToLower(signature)), []byte(expected)) != 1 {
		return nil, validationError("webhook signature is invalid")
	}
	return allQueryValues(query), nil
}

func buildWebhookSignatureDebug(headers http.Header, query url.Values, now time.Time, tolerance time.Duration, signedFields []string) WebhookSignatureDebug {
	timestamp := strings.TrimSpace(headers.Get(WebhookHeaderTimestamp))
	received := strings.TrimSpace(headers.Get(WebhookHeaderSignature))
	signedValues := make(map[string]string, len(signedFields))
	for _, field := range signedFields {
		signedValues[field] = query.Get(field)
	}
	signingString := callbackSigningString(timestamp, signedValues, signedFields)
	expected := sha256Hex(signingString)
	timestampValid := true
	var timestampError string
	if tolerance > 0 {
		if err := validateWebhookTimestamp(timestamp, now, tolerance); err != nil {
			timestampValid = false
			timestampError = err.Error()
		}
	}
	return WebhookSignatureDebug{
		Timestamp:        timestamp,
		ReceivedSign:     received,
		ExpectedSign:     expected,
		SigningString:    signingString,
		SignMatched:      received != "" && subtle.ConstantTimeCompare([]byte(strings.ToLower(received)), []byte(expected)) == 1,
		TimestampValid:   timestampValid,
		TimestampError:   timestampError,
		SignedFieldNames: append([]string(nil), signedFields...),
		SignedValues:     signedValues,
		QueryValues:      allQueryValues(query),
	}
}

func validateWebhookTimestamp(timestamp string, now time.Time, tolerance time.Duration) error {
	value, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return validationError("webhook header t must be millisecond timestamp")
	}
	callbackTime := time.UnixMilli(value)
	diff := now.Sub(callbackTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		return validationError("webhook timestamp is outside tolerance")
	}
	return nil
}

func requiredQueryValues(query url.Values, fields []string) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		value := query.Get(field)
		if strings.TrimSpace(value) == "" {
			return nil, validationError("webhook query " + field + " can not be blank")
		}
		values[field] = value
	}
	return values, nil
}

func allQueryValues(query url.Values) map[string]string {
	values := make(map[string]string, len(query))
	for key := range query {
		values[key] = query.Get(key)
	}
	return values
}

func callbackSignature(timestamp string, values map[string]string, fields []string) string {
	return sha256Hex(callbackSigningString(timestamp, values, fields))
}

func callbackSigningString(timestamp string, values map[string]string, fields []string) string {
	var builder strings.Builder
	builder.WriteString(timestamp)
	for _, field := range fields {
		builder.WriteString(values[field])
	}
	return builder.String()
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
