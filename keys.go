package paymentgateway

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"regexp"
	"strings"
)

var pemBlockPattern = regexp.MustCompile(`(?s)-----BEGIN (?:PUBLIC|PRIVATE|RSA PRIVATE) KEY-----.+?-----END (?:PUBLIC|PRIVATE|RSA PRIVATE) KEY-----`)

func ReadPublicKey(value string) (*rsa.PublicKey, error) {
	der, err := normalizePEM(value)
	if err != nil {
		return nil, cryptoError("OpenAPI platform public key can not be parsed", err)
	}
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, cryptoError("OpenAPI platform public key can not be parsed", err)
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, cryptoError("OpenAPI platform public key is not RSA", nil)
	}
	return rsaKey, nil
}

func ReadPrivateKey(value string) (*rsa.PrivateKey, error) {
	der, err := normalizePEM(value)
	if err != nil {
		return nil, cryptoError("OpenAPI merchant response private key can not be parsed", err)
	}
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, cryptoError("OpenAPI merchant response private key is not RSA", nil)
		}
		return rsaKey, nil
	}
	pkcs1Key, pkcs1Err := x509.ParsePKCS1PrivateKey(der)
	if pkcs1Err != nil {
		return nil, cryptoError("OpenAPI merchant response private key can not be parsed", err)
	}
	return pkcs1Key, nil
}

func normalizePEM(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, cryptoError("OpenAPI key can not be blank", nil)
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
