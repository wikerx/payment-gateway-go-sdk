package paymentgateway

// EncryptPayload encrypts a plain JSON request body into the OpenAPI compact
// payload format: protectedHeader.encryptedAesKey.iv.cipherText.tag.
//
// platformPublicKeyText supports either PEM text or X.509 DER Base64 text.
// This function does not read config files, sign JWTs, or send HTTP requests;
// it is provided as a small copy-friendly helper for merchants who want to
// verify the OpenAPI payload encryption protocol independently.
func EncryptPayload(plainJSON string, platformPublicKeyText string) (string, error) {
	publicKey, err := ReadPublicKey(platformPublicKeyText)
	if err != nil {
		return "", err
	}
	return NewPayloadCrypto().Encrypt(plainJSON, publicKey)
}

// DecryptPayload decrypts an OpenAPI compact payload into its plain JSON body.
//
// merchantResponsePrivateKeyText supports either PEM text or PKCS#8 DER Base64
// text. This function only decrypts the compact data value and does not parse
// the gateway response envelope code/msg/livemode.
func DecryptPayload(compactData string, merchantResponsePrivateKeyText string) (string, error) {
	privateKey, err := ReadPrivateKey(merchantResponsePrivateKeyText)
	if err != nil {
		return "", err
	}
	return NewPayloadCrypto().Decrypt(compactData, privateKey)
}
