package paymentgateway

import "fmt"

type ErrorKind string

const (
	ErrorKindConfig     ErrorKind = "config"
	ErrorKindCrypto     ErrorKind = "crypto"
	ErrorKindHTTP       ErrorKind = "http"
	ErrorKindResponse   ErrorKind = "response"
	ErrorKindValidation ErrorKind = "validation"
)

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
