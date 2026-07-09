package paymentgateway

import "fmt"

type ErrorKind string

const (
	// ErrorKindConfig indicates invalid or missing SDK configuration.
	ErrorKindConfig ErrorKind = "config"
	// ErrorKindCrypto indicates encryption, decryption, or key parsing failure.
	ErrorKindCrypto ErrorKind = "crypto"
	// ErrorKindHTTP indicates request construction, transport, or HTTP status failure.
	ErrorKindHTTP ErrorKind = "http"
	// ErrorKindResponse indicates the gateway response envelope or payload is invalid.
	ErrorKindResponse ErrorKind = "response"
	// ErrorKindValidation indicates merchant input validation failure before sending.
	ErrorKindValidation ErrorKind = "validation"
)

// SDKError is the typed error returned by SDK validation, transport, crypto,
// config, and response parsing paths.
type SDKError struct {
	Kind ErrorKind
	Msg  string
	Err  error
}

func (e *SDKError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return fmt.Sprintf("openapi %s error: %s", e.Kind, e.Msg)
	}
	return fmt.Sprintf("openapi %s error: %s: %v", e.Kind, e.Msg, e.Err)
}

func (e *SDKError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func configError(msg string, err error) error {
	return &SDKError{Kind: ErrorKindConfig, Msg: msg, Err: err}
}

func cryptoError(msg string, err error) error {
	return &SDKError{Kind: ErrorKindCrypto, Msg: msg, Err: err}
}

func httpError(msg string, err error) error {
	return &SDKError{Kind: ErrorKindHTTP, Msg: msg, Err: err}
}

func responseError(msg string, err error) error {
	return &SDKError{Kind: ErrorKindResponse, Msg: msg, Err: err}
}

func validationError(msg string) error {
	return &SDKError{Kind: ErrorKindValidation, Msg: msg}
}
