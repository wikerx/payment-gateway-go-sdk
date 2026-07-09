package payloadcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"regexp"
	"strings"
)

const (
	// payloadType identifies this compact payload as a Payment Gateway business
	// body. The gateway validates this value before decrypting the request.
	payloadType = "PAYMENT-PAYLOAD"

	// keyEncryptionAlg is the asymmetric algorithm used to protect the random
	// AES content key. It must match the gateway protocol exactly.
	keyEncryptionAlg = "RSA-OAEP-256"

	// contentEncryptionAlg is the symmetric payload encryption algorithm.
	// A256GCM means AES-256-GCM.
	contentEncryptionAlg = "A256GCM"

	// aesKeyBytes is 32 bytes because AES-256 requires a 256-bit content key.
	aesKeyBytes = 32

	// gcmIVBytes is the standard 96-bit nonce length recommended for GCM.
	gcmIVBytes = 12

	// Go's cipher.GCM appends a 16-byte authentication tag to the ciphertext.
	gcmTagBytes = 16

	// Compact payload format:
	// protectedHeader.encryptedAesKey.iv.cipherText.tag
	compactParts = 5
)

// pemBlockPattern extracts the first public/private key block when merchants
// paste a key together with surrounding comments or extra whitespace.
var pemBlockPattern = regexp.MustCompile(`(?s)-----BEGIN (?:PUBLIC|PRIVATE|RSA PRIVATE) KEY-----.+?-----END (?:PUBLIC|PRIVATE|RSA PRIVATE) KEY-----`)

// EncryptPayload encrypts a plain JSON request body into OpenAPI compact data:
// protectedHeader.encryptedAesKey.iv.cipherText.tag.
//
// platformPublicKeyText supports PEM text or X.509 DER Base64 text.
//
// This function only builds the encrypted data value used in the HTTP body:
//
//	{
//	  "livemode": false,
//	  "data": "<return value>"
//	}
//
// It does not generate JWT, send HTTP requests, read config files, or modify
// the business JSON. The caller should pass the final JSON string that needs to
// be encrypted.
func EncryptPayload(plainJSON string, platformPublicKeyText string) (string, error) {
	if strings.TrimSpace(plainJSON) == "" {
		return "", errors.New("plainJSON can not be blank")
	}

	// The request body must be encrypted with the platform request public key.
	// Only the gateway owns the matching private key, so only the gateway can
	// decrypt merchant requests.
	publicKey, err := readPublicKey(platformPublicKeyText)
	if err != nil {
		return "", err
	}

	// Each payload gets a new AES content key. Do not reuse it across requests.
	contentKey, err := randomBytes(aesKeyBytes)
	if err != nil {
		return "", err
	}

	// Each AES-GCM encryption also gets a fresh 12-byte IV/nonce.
	iv, err := randomBytes(gcmIVBytes)
	if err != nil {
		return "", err
	}

	// The protected header is Base64URL(JSON). It is both transmitted as the
	// first compact segment and authenticated as AES-GCM AAD.
	protectedHeader, err := protectedHeader()
	if err != nil {
		return "", err
	}

	// AES-GCM returns ciphertext followed by the authentication tag. The compact
	// protocol sends them as separate segments, so they are split below.
	cipherWithTag, err := aesGCMSeal(contentKey, iv, []byte(protectedHeader), []byte(plainJSON))
	if err != nil {
		return "", err
	}

	// The random AES content key is encrypted with RSA-OAEP-SHA256 so it can be
	// safely sent together with the ciphertext.
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, contentKey, nil)
	if err != nil {
		return "", err
	}

	cipherText := cipherWithTag[:len(cipherWithTag)-gcmTagBytes]
	tag := cipherWithTag[len(cipherWithTag)-gcmTagBytes:]

	// RawURLEncoding is Base64URL without "=" padding, which keeps the compact
	// data URL/header safe and consistent with the gateway protocol.
	return strings.Join([]string{
		protectedHeader,
		base64.RawURLEncoding.EncodeToString(encryptedKey),
		base64.RawURLEncoding.EncodeToString(iv),
		base64.RawURLEncoding.EncodeToString(cipherText),
		base64.RawURLEncoding.EncodeToString(tag),
	}, "."), nil
}

// DecryptPayload decrypts OpenAPI compact data into plain JSON.
//
// merchantResponsePrivateKeyText supports PEM text, PKCS#8 DER Base64 text,
// or PKCS#1 RSA private key PEM text.
//
// This function only decrypts the compact data value from the gateway response.
// The caller is still responsible for checking the response envelope, such as
// code/msg/livemode, before trusting the business result.
func DecryptPayload(compactData string, merchantResponsePrivateKeyText string) (string, error) {
	if strings.TrimSpace(compactData) == "" {
		return "", errors.New("compactData can not be blank")
	}

	// Gateway responses are encrypted for the merchant. The merchant response
	// private key is required to unwrap the AES content key.
	privateKey, err := readPrivateKey(merchantResponsePrivateKeyText)
	if err != nil {
		return "", err
	}

	parts := strings.Split(compactData, ".")
	if len(parts) != compactParts {
		return "", errors.New("compactData format is invalid")
	}

	// Validate the header before using it as AES-GCM AAD. If the header was
	// changed, AES-GCM authentication will fail later as well, but this gives a
	// clearer protocol-level error.
	if err := validateProtectedHeader(parts[0]); err != nil {
		return "", err
	}

	encryptedKey, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	// Recover the random AES content key from the RSA-OAEP encrypted segment.
	contentKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return "", err
	}

	iv, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", err
	}
	cipherText, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return "", err
	}
	tag, err := base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil {
		return "", err
	}

	// Go expects ciphertext and tag together when opening AES-GCM, while the
	// compact protocol carries them as two separate segments.
	plain, err := aesGCMOpen(contentKey, iv, []byte(parts[0]), append(cipherText, tag...))
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// protectedHeader builds the Base64URL protected header segment.
//
// The same encoded string is used as AES-GCM AAD, so the alg/enc/typ metadata
// is integrity-protected and cannot be changed without breaking decryption.
func protectedHeader() (string, error) {
	headerBytes, err := json.Marshal(map[string]string{
		"typ": payloadType,
		"alg": keyEncryptionAlg,
		"enc": contentEncryptionAlg,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(headerBytes), nil
}

// validateProtectedHeader checks that the compact payload declares exactly the
// encryption profile this gateway supports.
func validateProtectedHeader(value string) error {
	headerBytes, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return err
	}
	var header map[string]string
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return err
	}
	if header["typ"] != payloadType || header["alg"] != keyEncryptionAlg || header["enc"] != contentEncryptionAlg {
		return errors.New("compactData protected header is invalid")
	}
	return nil
}

// aesGCMSeal encrypts plainText with AES-256-GCM.
//
// aad is authenticated but not encrypted. Here it is the protected header
// segment, matching the compact payload protocol.
func aesGCMSeal(contentKey, iv, aad, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(contentKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nil, iv, plainText, aad), nil
}

// aesGCMOpen verifies the GCM tag and decrypts the ciphertext. Authentication
// failure returns an error and no plaintext.
func aesGCMOpen(contentKey, iv, aad, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(contentKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, iv, cipherText, aad)
}

// randomBytes returns cryptographically secure random bytes for AES keys and
// GCM IVs.
func randomBytes(length int) ([]byte, error) {
	value := make([]byte, length)
	_, err := rand.Read(value)
	return value, err
}

// readPublicKey accepts either:
//   - PEM text: -----BEGIN PUBLIC KEY----- ... -----END PUBLIC KEY-----
//   - DER Base64 text for an X.509 SubjectPublicKeyInfo public key
func readPublicKey(value string) (*rsa.PublicKey, error) {
	der, err := normalizeKey(value)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}
	return rsaKey, nil
}

// readPrivateKey accepts merchant response private keys in the common formats:
//   - PEM text: -----BEGIN PRIVATE KEY----- ... -----END PRIVATE KEY-----
//   - PKCS#8 DER Base64 text
//   - PKCS#1 RSA private key PEM text
func readPrivateKey(value string) (*rsa.PrivateKey, error) {
	der, err := normalizeKey(value)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
		return rsaKey, nil
	}
	rsaKey, pkcs1Err := x509.ParsePKCS1PrivateKey(der)
	if pkcs1Err != nil {
		return nil, err
	}
	return rsaKey, nil
}

// normalizeKey converts pasted key text into raw DER bytes. It tolerates extra
// newlines, spaces, tabs, and surrounding text so merchants can paste keys from
// config files or dashboards with minimal cleanup.
func normalizeKey(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("key can not be blank")
	}
	if match := pemBlockPattern.FindString(value); match != "" {
		value = match
	}
	if block, _ := pem.Decode([]byte(value)); block != nil {
		return block.Bytes, nil
	}
	cleaned := strings.NewReplacer("\n", "", "\r", "", "\t", "", " ", "").Replace(value)
	return base64.StdEncoding.DecodeString(cleaned)
}
