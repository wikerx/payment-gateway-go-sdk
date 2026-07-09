package paymentgateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"time"
)

type MerchantJWTSigner struct{}

func NewMerchantJWTSigner() *MerchantJWTSigner {
	return &MerchantJWTSigner{}
}

func (s *MerchantJWTSigner) Sign(merchantID, secret string, livemode bool, jwtID string, issuedAt time.Time, ttlSeconds int) (string, error) {
	if strings.TrimSpace(merchantID) == "" {
		return "", validationError("merchantId can not be blank")
	}
	if len([]byte(secret)) < 32 {
		return "", validationError("merchant jwt secret must be at least 256 bits for HS256")
	}
	if strings.TrimSpace(jwtID) == "" {
		return "", validationError("jwt jti can not be blank")
	}
	if ttlSeconds <= 0 || ttlSeconds > JWTTTLSeconds {
		return "", validationError("jwt ttlSeconds must be between 1 and 180")
	}
	header := map[string]string{
		"alg":         "HS256",
		JWTHeaderType: JWTType,
	}
	claims := map[string]any{
		"aud":        []string{"gateway"},
		"iss":        "merchant",
		"jti":        jwtID,
		"iat":        issuedAt.Unix(),
		"exp":        issuedAt.Add(time.Duration(ttlSeconds) * time.Second).Unix(),
		"merchantId": merchantID,
		"livemode":   livemode,
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := base64RawURL(headerJSON) + "." + base64RawURL(claimsJSON)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	return signingInput + "." + base64RawURL(mac.Sum(nil)), nil
}
