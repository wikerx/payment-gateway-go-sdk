package paymentgateway

import (
	"encoding/json"
	"strings"
)

type WebhookVerifier struct {
	payloadCrypto *PayloadCrypto
	privateKey    string
	livemode      bool
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
