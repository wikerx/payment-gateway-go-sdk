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
	payloadType          = "PAYMENT-PAYLOAD"
	keyEncryptionAlg     = "RSA-OAEP-256"
	contentEncryptionAlg = "A256GCM"
	aesKeyBytes          = 32
	gcmIVBytes           = 12
	gcmTagBytes          = 16
	compactParts         = 5
)

var pemBlockPattern = regexp.MustCompile(`(?s)-----BEGIN (?:PUBLIC|PRIVATE|RSA PRIVATE) KEY-----.+?-----END (?:PUBLIC|PRIVATE|RSA PRIVATE) KEY-----`)

// EncryptPayload encrypts a plain JSON request body into OpenAPI compact data:
// protectedHeader.encryptedAesKey.iv.cipherText.tag.
//
// platformPublicKeyText supports PEM text or X.509 DER Base64 text.
func EncryptPayload(plainJSON string, platformPublicKeyText string) (string, error) {
	if strings.TrimSpace(plainJSON) == "" {
		return "", errors.New("plainJSON can not be blank")
	}
	publicKey, err := readPublicKey(platformPublicKeyText)
	if err != nil {
		return "", err
	}
	contentKey, err := randomBytes(aesKeyBytes)
	if err != nil {
		return "", err
	}
	iv, err := randomBytes(gcmIVBytes)
	if err != nil {
		return "", err
	}
	protectedHeader, err := protectedHeader()
	if err != nil {
		return "", err
	}
	cipherWithTag, err := aesGCMSeal(contentKey, iv, []byte(protectedHeader), []byte(plainJSON))
	if err != nil {
		return "", err
	}
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, contentKey, nil)
	if err != nil {
		return "", err
	}
	cipherText := cipherWithTag[:len(cipherWithTag)-gcmTagBytes]
	tag := cipherWithTag[len(cipherWithTag)-gcmTagBytes:]
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
func DecryptPayload(compactData string, merchantResponsePrivateKeyText string) (string, error) {
	if strings.TrimSpace(compactData) == "" {
		return "", errors.New("compactData can not be blank")
	}
	privateKey, err := readPrivateKey(merchantResponsePrivateKeyText)
	if err != nil {
		return "", err
	}
	parts := strings.Split(compactData, ".")
	if len(parts) != compactParts {
		return "", errors.New("compactData format is invalid")
	}
	if err := validateProtectedHeader(parts[0]); err != nil {
		return "", err
	}
	encryptedKey, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
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
	plain, err := aesGCMOpen(contentKey, iv, []byte(parts[0]), append(cipherText, tag...))
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

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

func randomBytes(length int) ([]byte, error) {
	value := make([]byte, length)
	_, err := rand.Read(value)
	return value, err
}

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
