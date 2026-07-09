package payloadcrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestPayloadCryptoStandaloneWithPEMText(t *testing.T) {
	publicText, privateText := testKeyPairText(t, true)
	compact, err := EncryptPayload(`{"orderNo":"PAYIN_1","amount":"12.34"}`, publicText)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := DecryptPayload(compact, privateText)
	if err != nil {
		t.Fatal(err)
	}
	if plain != `{"orderNo":"PAYIN_1","amount":"12.34"}` {
		t.Fatalf("unexpected plain json: %s", plain)
	}
}

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
