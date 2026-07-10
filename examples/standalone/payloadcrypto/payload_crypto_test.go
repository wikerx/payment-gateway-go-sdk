package payloadcrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
	"time"
)

// TestPayloadCryptoStandaloneWithPEMText verifies that merchants can use PEM
// public/private key text with EncryptPayload and DecryptPayload.
func TestPayloadCryptoStandaloneWithPEMText(t *testing.T) {
	publicText, privateText := testKeyPairText(t, true)
	plainJSON := `{"message":"standalone payload crypto round trip","amount":"12.34"}`
	compact, err := EncryptPayload(plainJSON, publicText)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := DecryptPayload(compact, privateText)
	if err != nil {
		t.Fatal(err)
	}
	if plain != plainJSON {
		t.Fatalf("unexpected plain json: %s", plain)
	}
}

// TestPayloadCryptoStandaloneWithDERBase64Text verifies the same payload flow
// using DER Base64 key text, matching the inline key style in merchant config.
func TestPayloadCryptoStandaloneWithDERBase64Text(t *testing.T) {
	publicText, privateText := testKeyPairText(t, false)
	compact, err := EncryptPayload(`{"currency":"USD"}`, publicText)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := DecryptPayload(compact, privateText)
	if err != nil {
		t.Fatal(err)
	}
	if plain != `{"currency":"USD"}` {
		t.Fatalf("unexpected plain json: %s", plain)
	}
}

// TestBuildOpenAPIHeaders verifies that the standalone helper creates the
// required OpenAPI headers and embeds merchant identity into the JWT claims.
func TestBuildOpenAPIHeaders(t *testing.T) {
	headers, err := BuildOpenAPIHeaders("2606177036", strings.Repeat("s", 32), false, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(headers["Authorization"], "Bearer ") {
		t.Fatalf("authorization header missing bearer token: %#v", headers)
	}
	if headers["Content-Type"] != "application/json; charset=UTF-8" {
		t.Fatalf("content type missing: %#v", headers)
	}
	if headers["Accept"] != "application/json" || headers["User-Agent"] == "" || headers["X-Request-Id"] == "" {
		t.Fatalf("required headers missing: %#v", headers)
	}

	claims := decodeJWTClaims(t, strings.TrimPrefix(headers["Authorization"], "Bearer "))
	if claims["merchantId"] != "2606177036" {
		t.Fatalf("merchantId claim mismatch: %#v", claims)
	}
	if claims["livemode"] != false {
		t.Fatalf("livemode claim mismatch: %#v", claims)
	}
	if strings.TrimSpace(claims["jti"].(string)) == "" {
		t.Fatalf("jti claim missing: %#v", claims)
	}
}

// TestBuildOpenAPIHeadersGeneratesFreshJWTID protects the replay requirement:
// every call must create a fresh JWT jti and therefore a different token.
func TestBuildOpenAPIHeadersGeneratesFreshJWTID(t *testing.T) {
	first, err := BuildOpenAPIHeaders("2606177036", strings.Repeat("s", 32), false, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildOpenAPIHeaders("2606177036", strings.Repeat("s", 32), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if first["Authorization"] == second["Authorization"] {
		t.Fatal("authorization JWT must be unique per request")
	}
	if _, ok := first["Content-Type"]; ok {
		t.Fatalf("content type should be omitted when withBody=false: %#v", first)
	}
}

// TestSignMerchantJWT verifies deterministic JWT claims when callers choose to
// build Authorization headers themselves.
func TestSignMerchantJWT(t *testing.T) {
	token, err := SignMerchantJWT("2606177036", strings.Repeat("s", 32), true, "JTI_1", time.Unix(1000, 0).UTC(), 180)
	if err != nil {
		t.Fatal(err)
	}
	claims := decodeJWTClaims(t, token)
	if claims["jti"] != "JTI_1" || claims["merchantId"] != "2606177036" || claims["livemode"] != true {
		t.Fatalf("claims mismatch: %#v", claims)
	}
	if claims["iat"].(float64) != 1000 || claims["exp"].(float64) != 1180 {
		t.Fatalf("time claims mismatch: %#v", claims)
	}
}

// testKeyPairText creates a temporary RSA key pair and returns it as PEM text
// or DER Base64 text for standalone crypto tests.
func testKeyPairText(t *testing.T, pemText bool) (string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if !pemText {
		return base64.StdEncoding.EncodeToString(publicDER), base64.StdEncoding.EncodeToString(privateDER)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}))
}

// decodeJWTClaims decodes the unsigned JWT payload segment for test assertions.
func decodeJWTClaims(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid jwt format: %s", token)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	return claims
}
