package paymentgateway

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"strings"
	"testing"
)

type mockTransport struct {
	request SDKHTTPRequest
	body    string
}

func (m *mockTransport) Execute(ctx context.Context, request SDKHTTPRequest) (*SDKHTTPResponse, error) {
	m.request = request
	return &SDKHTTPResponse{StatusCode: 200, Body: m.body}, nil
}

func TestClientCreateCustomerUsesEncryptedBody(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	cfg := testConfig(t, privateKey)
	crypto := NewPayloadCrypto()
	responseData, err := crypto.Encrypt(`{"customerId":"cus_1"}`, &privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	livemode := false
	transport := &mockTransport{body: `{"code":0,"msg":"success","livemode":false,"data":"` + responseData + `"}`}
	client, err := NewClient(cfg, WithHTTPTransport(transport))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.CreateCustomer(context.Background(), APIRequest{
		"firstname": "Lily",
		"lastname":  "Brown",
		"email":     "lily@example.com",
		"country":   "US",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Livemode == nil || *result.Livemode != livemode {
		t.Fatal("livemode not parsed")
	}
	if !strings.Contains(transport.request.Headers[HeaderAuthorization], AuthorizationPrefix) {
		t.Fatal("authorization header missing")
	}
	if !strings.Contains(transport.request.Body, `"livemode":false`) || !strings.Contains(transport.request.Body, `"data"`) {
		t.Fatalf("encrypted request body not generated: %s", transport.request.Body)
	}
}

func TestClientLogsEncryptedResponseAndPayloadParts(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	cfg := testConfig(t, privateKey)
	cfg.RawHTTPLogEnabled = true
	crypto := NewPayloadCrypto()
	responseData, err := crypto.Encrypt(`{"customerId":"cus_1"}`, &privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	transport := &mockTransport{body: `{"code":0,"msg":"success","livemode":false,"data":"` + responseData + `"}`}
	client, err := NewClient(cfg, WithHTTPTransport(transport))
	if err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(originalOutput)
	_, err = client.CreateCustomer(context.Background(), APIRequest{
		"firstname": "Lily",
		"lastname":  "Brown",
		"email":     "lily@example.com",
		"country":   "US",
	})
	if err != nil {
		t.Fatal(err)
	}
	output := logs.String()
	if !strings.Contains(output, "响应原始密文参数") {
		t.Fatalf("encrypted response log missing: %s", output)
	}
	if !strings.Contains(output, "响应参数拆分") {
		t.Fatalf("response payload parts log missing: %s", output)
	}
}

func testConfig(t *testing.T, key *rsa.PrivateKey) *Config {
	t.Helper()
	privateDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return &Config{
		BaseURL:                    "http://localhost:58060",
		MerchantID:                 "2606177036",
		MerchantJWTSecret:          strings.Repeat("a", 32),
		Livemode:                   false,
		PlatformPublicKey:          string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})),
		MerchantResponsePrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})),
	}
}
