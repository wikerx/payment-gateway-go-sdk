package paymentgateway

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestPayloadCryptoRoundTrip(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	crypto := NewPayloadCrypto()
	compact, err := crypto.Encrypt(`{"orderNo":"ORD_1"}`, &privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := crypto.Decrypt(compact, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	if plain != `{"orderNo":"ORD_1"}` {
		t.Fatalf("unexpected plain text: %s", plain)
	}
	parts, err := crypto.SplitCompactPayload(compact)
	if err != nil {
		t.Fatal(err)
	}
	if parts.Header == "" || parts.EncryptedAESKey == "" || parts.Tag == "" {
		t.Fatalf("payload parts not populated: %#v", parts)
	}
}

func TestReadRSAKeysFromPEM(t *testing.T) {
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
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}))
	publicPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}))
	if _, err := ReadPrivateKey(privatePEM); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPublicKey(publicPEM); err != nil {
		t.Fatal(err)
	}
}
