package paymentgateway

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestStandalonePayloadCryptoWithPEMText(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicText, privateText := testKeyText(t, privateKey, true)

	compact, err := EncryptPayload(`{"amount":"12.34"}`, publicText)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := DecryptPayload(compact, privateText)
	if err != nil {
		t.Fatal(err)
	}
	if plain != `{"amount":"12.34"}` {
		t.Fatalf("unexpected decrypted payload: %s", plain)
	}
}

func TestStandalonePayloadCryptoWithDERBase64Text(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicText, privateText := testKeyText(t, privateKey, false)

	compact, err := EncryptPayload(`{"currency":"USD"}`, publicText)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := DecryptPayload(compact, privateText)
	if err != nil {
		t.Fatal(err)
	}
	if plain != `{"currency":"USD"}` {
		t.Fatalf("unexpected decrypted payload: %s", plain)
	}
}

func testKeyText(t *testing.T, privateKey *rsa.PrivateKey, pemText bool) (string, string) {
	t.Helper()
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
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	return string(publicPEM), string(privatePEM)
}
