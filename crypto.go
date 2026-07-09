package paymentgateway

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

const (
	aesKeyBytes = 32
	gcmIVBytes  = 12
	gcmTagBytes = 16
	partsCount  = 5
)

type PayloadParts struct {
	// ProtectedHeader is Base64URL(JSON) containing typ/alg/enc.
	ProtectedHeader string `json:"protectedHeader"`
	// Header is the decoded protected header JSON, useful for debug logs.
	Header string `json:"header"`
	// EncryptedAESKey is the RSA-OAEP-256 encrypted random AES content key.
	EncryptedAESKey string `json:"encryptedAesKey"`
	// IV is the AES-GCM nonce.
	IV string `json:"iv"`
	// CipherText is the AES-GCM encrypted business JSON without the tag.
	CipherText string `json:"cipherText"`
	// Tag is the AES-GCM authentication tag.
	Tag string `json:"tag"`
}

// CompactPayload joins payload components into the OpenAPI data format:
// protectedHeader.encryptedAesKey.iv.cipherText.tag.
func (p PayloadParts) CompactPayload() string {
	return strings.Join([]string{p.ProtectedHeader, p.EncryptedAESKey, p.IV, p.CipherText, p.Tag}, ".")
}

// PayloadCrypto implements the gateway compact payload encryption protocol. It
// is stateless and safe to reuse across requests.
type PayloadCrypto struct{}

// NewPayloadCrypto returns a stateless payload crypto helper.
func NewPayloadCrypto() *PayloadCrypto {
	return &PayloadCrypto{}
}

// Encrypt encrypts plainText with a random AES-256-GCM content key and encrypts
// that content key with recipientPublicKey using RSA-OAEP-SHA256.
func (c *PayloadCrypto) Encrypt(plainText string, recipientPublicKey *rsa.PublicKey) (string, error) {
	parts, err := c.EncryptToParts(plainText, recipientPublicKey)
	if err != nil {
		return "", err
	}
	return parts.CompactPayload(), nil
}

// EncryptToParts encrypts plainText and returns the individual compact payload
// segments. It is mainly useful for debug logging and protocol tests.
func (c *PayloadCrypto) EncryptToParts(plainText string, recipientPublicKey *rsa.PublicKey) (*PayloadParts, error) {
	if plainText == "" {
		return nil, cryptoError("plainText can not be blank", nil)
	}
	if recipientPublicKey == nil {
		return nil, cryptoError("recipientPublicKey can not be nil", nil)
	}
	contentKey, err := randomBytes(aesKeyBytes)
	if err != nil {
		return nil, cryptoError("OpenAPI random AES key failed", err)
	}
	iv, err := randomBytes(gcmIVBytes)
	if err != nil {
		return nil, cryptoError("OpenAPI random IV failed", err)
	}
	protectedHeader, headerJSON, err := encodeProtectedHeader()
	if err != nil {
		return nil, err
	}
	// The protected header is used as AAD so typ/alg/enc are integrity-protected
	// even though the header itself is not encrypted.
	cipherWithTag, err := aesGCMSeal(contentKey, iv, []byte(protectedHeader), []byte(plainText))
	if err != nil {
		return nil, err
	}
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, recipientPublicKey, contentKey, nil)
	if err != nil {
		return nil, cryptoError("OpenAPI RSA-OAEP crypto failed", err)
	}
	cipherText := cipherWithTag[:len(cipherWithTag)-gcmTagBytes]
	tag := cipherWithTag[len(cipherWithTag)-gcmTagBytes:]
	return &PayloadParts{
		ProtectedHeader: protectedHeader,
		Header:          headerJSON,
		EncryptedAESKey: base64RawURL(encryptedKey),
		IV:              base64RawURL(iv),
		CipherText:      base64RawURL(cipherText),
		Tag:             base64RawURL(tag),
	}, nil
}

// Decrypt decrypts the compact payload data value returned by the gateway.
// privateKey must be the merchant response private key.
func (c *PayloadCrypto) Decrypt(compactPayload string, privateKey *rsa.PrivateKey) (string, error) {
	if strings.TrimSpace(compactPayload) == "" {
		return "", cryptoError("OpenAPI encrypted data can not be blank", nil)
	}
	if privateKey == nil {
		return "", cryptoError("privateKey can not be nil", nil)
	}
	parts, err := c.SplitCompactPayload(compactPayload)
	if err != nil {
		return "", err
	}
	encryptedKey, err := base64RawURLDecode(parts.EncryptedAESKey)
	if err != nil {
		return "", cryptoError("OpenAPI base64url data can not be decoded", err)
	}
	contentKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return "", cryptoError("OpenAPI RSA-OAEP crypto failed", err)
	}
	iv, err := base64RawURLDecode(parts.IV)
	if err != nil {
		return "", cryptoError("OpenAPI base64url data can not be decoded", err)
	}
	cipherText, err := base64RawURLDecode(parts.CipherText)
	if err != nil {
		return "", cryptoError("OpenAPI base64url data can not be decoded", err)
	}
	tag, err := base64RawURLDecode(parts.Tag)
	if err != nil {
		return "", cryptoError("OpenAPI base64url data can not be decoded", err)
	}
	plain, err := aesGCMOpen(contentKey, iv, []byte(parts.ProtectedHeader), append(cipherText, tag...))
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// SplitCompactPayload parses compact payload data into its five protocol
// segments and validates the protected header.
func (c *PayloadCrypto) SplitCompactPayload(compactPayload string) (*PayloadParts, error) {
	if strings.TrimSpace(compactPayload) == "" {
		return nil, cryptoError("OpenAPI encrypted data can not be blank", nil)
	}
	parts := strings.Split(compactPayload, ".")
	if len(parts) != partsCount {
		return nil, cryptoError("OpenAPI encrypted data format is invalid", nil)
	}
	headerJSON, err := decodeProtectedHeader(parts[0])
	if err != nil {
		return nil, err
	}
	return &PayloadParts{
		ProtectedHeader: parts[0],
		Header:          headerJSON,
		EncryptedAESKey: parts[1],
		IV:              parts[2],
		CipherText:      parts[3],
		Tag:             parts[4],
	}, nil
}

func encodeProtectedHeader() (string, string, error) {
	header := map[string]string{
		PayloadHeaderType: PayloadType,
		PayloadHeaderAlg:  KeyEncryptionAlg,
		PayloadHeaderEnc:  ContentEncryptionAlg,
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", "", cryptoError("OpenAPI encrypted data header can not be encoded", err)
	}
	return base64RawURL(headerBytes), string(headerBytes), nil
}

func decodeProtectedHeader(protectedHeader string) (string, error) {
	headerBytes, err := base64RawURLDecode(protectedHeader)
	if err != nil {
		return "", cryptoError("OpenAPI encrypted data header can not be parsed", err)
	}
	var header map[string]string
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", cryptoError("OpenAPI encrypted data header can not be parsed", err)
	}
	if header[PayloadHeaderType] != PayloadType || header[PayloadHeaderAlg] != KeyEncryptionAlg || header[PayloadHeaderEnc] != ContentEncryptionAlg {
		return "", cryptoError("OpenAPI encrypted data header is invalid", nil)
	}
	return string(headerBytes), nil
}

func aesGCMSeal(contentKey, iv, aad, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(contentKey)
	if err != nil {
		return nil, cryptoError("OpenAPI AES-GCM crypto failed", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, cryptoError("OpenAPI AES-GCM crypto failed", err)
	}
	return gcm.Seal(nil, iv, plainText, aad), nil
}

func aesGCMOpen(contentKey, iv, aad, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(contentKey)
	if err != nil {
		return nil, cryptoError("OpenAPI AES-GCM crypto failed", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, cryptoError("OpenAPI AES-GCM crypto failed", err)
	}
	plain, err := gcm.Open(nil, iv, cipherText, aad)
	if err != nil {
		return nil, cryptoError("OpenAPI AES-GCM crypto failed", err)
	}
	return plain, nil
}

func randomBytes(length int) ([]byte, error) {
	value := make([]byte, length)
	_, err := rand.Read(value)
	return value, err
}

func base64RawURL(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func base64RawURLDecode(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(value)
}
