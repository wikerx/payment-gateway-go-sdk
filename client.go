package paymentgateway

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	config                     *Config
	transport                  HTTPTransport
	payloadCrypto              *PayloadCrypto
	jwtSigner                  *MerchantJWTSigner
	platformPublicKey          *rsa.PublicKey
	merchantResponsePrivateKey *rsa.PrivateKey
}

// NewClient creates a Payment Gateway OpenAPI client from an already loaded
// Config. It validates merchant settings, parses RSA keys, and prepares the
// JWT signer and payload crypto helpers used by subsequent API calls.
func NewClient(config *Config, options ...ClientOption) (*Client, error) {
	if config == nil {
		return nil, configError("config can not be nil", nil)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	platformPublicKey, err := ReadPublicKey(config.PlatformPublicKey)
	if err != nil {
		return nil, err
	}
	merchantPrivateKey, err := ReadPrivateKey(config.MerchantResponsePrivateKey)
	if err != nil {
		return nil, err
	}
	client := &Client{
		config:                     config,
		transport:                  NewNetHTTPTransport(),
		payloadCrypto:              NewPayloadCrypto(),
		jwtSigner:                  NewMerchantJWTSigner(),
		platformPublicKey:          platformPublicKey,
		merchantResponsePrivateKey: merchantPrivateKey,
	}
	for _, option := range options {
		option(client)
	}
	return client, nil
}

// Create loads merchant-config.properties from configPath and returns a ready
// OpenAPI client. When configPath is blank, config/merchant-config.properties
// under the current working directory is used.
func Create(configPath string, options ...ClientOption) (*Client, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return NewClient(cfg, options...)
}

// ClientOption customizes a Client during construction.
type ClientOption func(*Client)

// WithHTTPTransport overrides the default net/http transport. Tests and
// merchants with their own HTTP stack can use this hook without changing
// encryption, JWT signing, or response parsing behavior.
func WithHTTPTransport(transport HTTPTransport) ClientOption {
	return func(client *Client) {
		if transport != nil {
			client.transport = transport
		}
	}
}

// Config returns the validated configuration used by the client.
func (c *Client) Config() *Config {
	return c.config
}

// CreateCheckoutPayment creates a hosted checkout payment order. The request is
// encrypted before being sent to /pay-api/trade/payment.
func (c *Client) CreateCheckoutPayment(ctx context.Context, request APIRequest) (*Result, error) {
	return c.createPayment(ctx, request)
}

// CreateLocalPayment creates a direct/local payment order. The paymentMethod
// and paymentMethodData fields should match the selected payment method.
func (c *Client) CreateLocalPayment(ctx context.Context, request APIRequest) (*Result, error) {
	return c.createPayment(ctx, request)
}

// CreateCardPayment creates a direct card payment order. If paymentMethod is
// blank, the SDK fills it with CARD before encrypting the request.
func (c *Client) CreateCardPayment(ctx context.Context, request APIRequest) (*Result, error) {
	if request == nil {
		return nil, validationError("payment request can not be nil")
	}
	if strings.TrimSpace(fmt.Sprint(request["paymentMethod"])) == "" {
		request["paymentMethod"] = PaymentMethodCard
	}
	return c.createPayment(ctx, request)
}

// RetrievePayment queries a payment order by gateway tradeNo.
func (c *Client) RetrievePayment(ctx context.Context, tradeNo string) (*Result, error) {
	value, err := requireText(tradeNo, "tradeNo")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointPaymentRetrieve, uniqueJWTID("PAYMENT_QUERY_"), url.PathEscape(value))
}

// CreateRefund creates a refund request for an existing payment. The request is
// encrypted and should include tradeNo, currency, amount, and refundAmount.
func (c *Client) CreateRefund(ctx context.Context, request APIRequest) (*Result, error) {
	if err := validateRefundCreateRequest(request); err != nil {
		return nil, err
	}
	return c.postEncrypted(ctx, EndpointRefundCreate, request, uniqueJWTID("REFUND_CREATE_"))
}

// RetrieveRefund queries a refund request by refundNo.
func (c *Client) RetrieveRefund(ctx context.Context, refundNo string) (*Result, error) {
	value, err := requireText(refundNo, "refundNo")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointRefundRetrieve, uniqueJWTID("REFUND_QUERY_"), url.PathEscape(value))
}

// CreatePayout creates a payout transfer. The request is encrypted and should
// include orderNo, currency, amount, paymentMethod, and paymentMethodData.
func (c *Client) CreatePayout(ctx context.Context, request APIRequest) (*Result, error) {
	if err := validatePayoutCreateRequest(request); err != nil {
		return nil, err
	}
	return c.postEncrypted(ctx, EndpointPayoutCreate, request, uniqueJWTID("PAYOUT_CREATE_"))
}

// RetrievePayout queries a payout transfer by gateway tradeNo.
func (c *Client) RetrievePayout(ctx context.Context, tradeNo string) (*Result, error) {
	value, err := requireText(tradeNo, "tradeNo")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointPayoutRetrieve, uniqueJWTID("PAYOUT_QUERY_"), url.PathEscape(value))
}

// CancelPayout cancels a payout transfer when the gateway still allows
// cancellation for the transfer state.
func (c *Client) CancelPayout(ctx context.Context, request APIRequest) (*Result, error) {
	if request == nil {
		return nil, validationError("request can not be nil")
	}
	return c.postEncrypted(ctx, EndpointPayoutCancel, request, uniqueJWTID("PAYOUT_CANCEL_"))
}

// RetrieveBalances queries merchant fund account balances. Passing a non-blank
// currency adds ?currency=<value>; passing blank queries all available balances.
func (c *Client) RetrieveBalances(ctx context.Context, currency string) (*Result, error) {
	path := EndpointBalanceInquiry.Path
	if strings.TrimSpace(currency) != "" {
		path += "?currency=" + url.QueryEscape(strings.TrimSpace(currency))
	}
	return c.execute(ctx, EndpointBalanceInquiry, path, nil, nil, uniqueJWTID("BALANCE_QUERY_"))
}

// CreateCustomer creates a customer profile that can be referenced by later
// payment requests through customerId.
func (c *Client) CreateCustomer(ctx context.Context, request APIRequest) (*Result, error) {
	if err := validateCustomerRequest(request); err != nil {
		return nil, err
	}
	return c.postEncrypted(ctx, EndpointCustomerCreate, request, uniqueJWTID("CUSTOMER_CREATE_"))
}

// RetrieveCustomer queries a customer profile by customerId.
func (c *Client) RetrieveCustomer(ctx context.Context, customerID string) (*Result, error) {
	value, err := requireText(customerID, "customerId")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointCustomerRetrieve, uniqueJWTID("CUSTOMER_QUERY_"), url.PathEscape(value))
}

// UpdateCustomer updates a customer profile by customerId. If customerId also
// exists in request, the path value wins and the body field is removed.
func (c *Client) UpdateCustomer(ctx context.Context, customerID string, request APIRequest) (*Result, error) {
	if err := validateCustomerRequest(request); err != nil {
		return nil, err
	}
	value, err := requireText(customerID, "customerId")
	if err != nil {
		return nil, err
	}
	delete(request, "customerId")
	path := EndpointCustomerUpdate.Format(url.PathEscape(value))
	return c.sendEncrypted(ctx, EndpointCustomerUpdate, path, request, uniqueJWTID("CUSTOMER_UPDATE_"))
}

// DeleteCustomer deletes or disables a customer profile by customerId,
// depending on gateway-side customer lifecycle rules.
func (c *Client) DeleteCustomer(ctx context.Context, customerID string) (*Result, error) {
	value, err := requireText(customerID, "customerId")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointCustomerDelete, uniqueJWTID("CUSTOMER_DELETE_"), url.PathEscape(value))
}

// ListCustomers returns the merchant customer list from the gateway.
func (c *Client) ListCustomers(ctx context.Context) (*Result, error) {
	return c.execute(ctx, EndpointCustomerList, EndpointCustomerList.Path, nil, nil, uniqueJWTID("CUSTOMER_LIST_"))
}

func (c *Client) createPayment(ctx context.Context, request APIRequest) (*Result, error) {
	if err := validatePaymentCreateRequest(request); err != nil {
		return nil, err
	}
	return c.postEncrypted(ctx, EndpointPaymentCreate, request, uniqueJWTID("PAYMENT_CREATE_"))
}

func (c *Client) postEncrypted(ctx context.Context, endpoint Endpoint, request APIRequest, jwtID string) (*Result, error) {
	return c.sendEncrypted(ctx, endpoint, endpoint.Path, request, jwtID)
}

func (c *Client) sendEncrypted(ctx context.Context, endpoint Endpoint, path string, request APIRequest, jwtID string) (*Result, error) {
	if request == nil {
		return nil, validationError("request can not be nil")
	}
	requestJSON, err := toJSON(request)
	if err != nil {
		return nil, validationError("request can not be serialized")
	}
	// Business request bodies are never sent as plain JSON. The SDK encrypts
	// them into compact payload data and wraps them with livemode.
	encryptedData, err := c.payloadCrypto.Encrypt(requestJSON, c.platformPublicKey)
	if err != nil {
		return nil, err
	}
	encryptedRequest := &EncryptedRequest{Livemode: c.config.Livemode, Data: encryptedData}
	return c.execute(ctx, endpoint, path, request, encryptedRequest, jwtID)
}

func (c *Client) getSecured(ctx context.Context, endpoint Endpoint, jwtID string, pathArgs ...any) (*Result, error) {
	return c.execute(ctx, endpoint, endpoint.Format(pathArgs...), nil, nil, jwtID)
}

func (c *Client) execute(ctx context.Context, endpoint Endpoint, path string, plainRequest APIRequest, encryptedRequest *EncryptedRequest, jwtID string) (*Result, error) {
	requestID := GenerateOrderNo("REQ_")
	requestURL := c.config.BaseURL + normalizePath(path)
	var body string
	var err error
	if encryptedRequest != nil {
		body, err = toJSON(encryptedRequest)
		if err != nil {
			return nil, err
		}
	}
	headers, err := c.headers(jwtID, requestID, body != "")
	if err != nil {
		return nil, err
	}
	// Every request carries a fresh JWT jti so the gateway can reject replayed
	// Authorization tokens independently from business order idempotency.
	c.logAPICall(endpoint, path, requestID, jwtID, plainRequest, encryptedRequest)
	response, err := c.transport.Execute(ctx, SDKHTTPRequest{
		Method:         endpoint.Method,
		URL:            requestURL,
		Headers:        headers,
		Body:           body,
		ConnectTimeout: c.config.ConnectTimeout,
		ReadTimeout:    c.config.ReadTimeout,
	})
	if err != nil {
		return nil, err
	}
	var encryptedResponse EncryptedResponse
	if err := fromJSON(response.Body, &encryptedResponse); err != nil {
		return nil, responseError("OpenAPI encrypted response can not be parsed", err)
	}
	c.logEncryptedResponse(response, &encryptedResponse)
	c.logResponsePayloadComponents(&encryptedResponse)
	if response.StatusCode < HTTPStatusSuccessMin || response.StatusCode >= HTTPStatusSuccessMax {
		return nil, httpError(fmt.Sprintf("OpenAPI HTTP status is not successful: %d", response.StatusCode), nil)
	}
	result, err := c.convertResult(&encryptedResponse, requestID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) convertResult(encryptedResponse *EncryptedResponse, requestID string) (*Result, error) {
	if encryptedResponse.Livemode != nil && *encryptedResponse.Livemode != c.config.Livemode {
		return nil, responseError("OpenAPI response livemode is inconsistent", nil)
	}
	result := &Result{
		Code:      encryptedResponse.Code,
		Msg:       encryptedResponse.Msg,
		Livemode:  encryptedResponse.Livemode,
		RequestID: requestID,
	}
	if strings.TrimSpace(encryptedResponse.Data) != "" {
		// Successful and error envelopes may both contain encrypted data. Always
		// decrypt data before exposing the final Result to merchant code.
		plainJSON, err := c.payloadCrypto.Decrypt(encryptedResponse.Data, c.merchantResponsePrivateKey)
		if err != nil {
			return nil, err
		}
		var data any
		if err := fromJSON(plainJSON, &data); err != nil {
			return nil, responseError("OpenAPI response data can not be parsed", err)
		}
		result.Data = data
		if c.config.RawHTTPLogEnabled {
			log.Printf("响应原始明文参数: %s", toPrettyJSON(Sanitize(data)))
		}
	}
	return result, nil
}

func (c *Client) headers(jwtID, requestID string, withBody bool) (map[string]string, error) {
	token, err := c.jwtSigner.Sign(c.config.MerchantID, c.config.MerchantJWTSecret, c.config.Livemode, jwtID, time.Now().UTC(), c.config.JWTTTLSeconds)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		HeaderAuthorization: AuthorizationPrefix + token,
		HeaderAccept:        Accept,
		HeaderUserAgent:     UserAgent,
		HeaderRequestID:     requestID,
	}
	if withBody {
		headers[HeaderContentType] = ContentType
	}
	return headers, nil
}

func (c *Client) logAPICall(endpoint Endpoint, path, requestID, jwtID string, plainRequest APIRequest, encryptedRequest *EncryptedRequest) {
	log.Printf("API调用开始: %s", toPrettyJSON(map[string]any{
		"apiName":    endpoint.Name,
		"method":     endpoint.Method,
		"path":       path,
		"merchantId": c.config.MerchantID,
		"requestId":  requestID,
		"jwtId":      jwtID,
	}))
	if c.config.RawHTTPLogEnabled && plainRequest != nil {
		log.Printf("请求原始明文报文: %s", toPrettyJSON(Sanitize(map[string]any(plainRequest))))
	}
	if encryptedRequest != nil {
		logData := map[string]any{"livemode": encryptedRequest.Livemode, "data": EncryptedDataSummary(encryptedRequest.Data)}
		if c.config.RawHTTPLogEnabled {
			logData["data"] = encryptedRequest.Data
		}
		log.Printf("请求密文参数: %s", toPrettyJSON(logData))
	}
}

func (c *Client) logEncryptedResponse(response *SDKHTTPResponse, encryptedResponse *EncryptedResponse) {
	if !c.config.RawHTTPLogEnabled || encryptedResponse == nil {
		return
	}
	log.Printf("响应原始密文参数: %s", toPrettyJSON(map[string]any{
		"statusCode": response.StatusCode,
		"headers":    SanitizeHeaders(response.Headers),
		"body":       encryptedResponse,
	}))
}

func (c *Client) logResponsePayloadComponents(encryptedResponse *EncryptedResponse) {
	if !c.config.RawHTTPLogEnabled || encryptedResponse == nil || strings.TrimSpace(encryptedResponse.Data) == "" {
		return
	}
	parts, err := c.payloadCrypto.SplitCompactPayload(encryptedResponse.Data)
	if err != nil {
		log.Printf("响应参数拆分跳过: %s", toPrettyJSON(map[string]any{"reason": "invalid_compact_payload"}))
		return
	}
	log.Printf("响应参数拆分: %s", toPrettyJSON(map[string]any{
		"protectedHeader": parts.ProtectedHeader,
		"header":          parts.Header,
		"encryptedAesKey": parts.EncryptedAESKey,
		"iv":              parts.IV,
		"cipherText":      parts.CipherText,
		"tag":             parts.Tag,
	}))
}

func normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func requireText(value string, fieldName string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", validationError(fieldName + " can not be blank")
	}
	return value, nil
}

func validatePaymentCreateRequest(request APIRequest) error {
	if request == nil {
		return validationError("payment request can not be nil")
	}
	for _, field := range []string{"orderNo", "currency", "amount"} {
		if strings.TrimSpace(fmt.Sprint(request[field])) == "" || fmt.Sprint(request[field]) == "<nil>" {
			return validationError(field + " can not be blank")
		}
	}
	return nil
}

func validateRefundCreateRequest(request APIRequest) error {
	if request == nil {
		return validationError("refund request can not be nil")
	}
	for _, field := range []string{"tradeNo", "currency", "amount", "refundAmount"} {
		if strings.TrimSpace(fmt.Sprint(request[field])) == "" || fmt.Sprint(request[field]) == "<nil>" {
			return validationError(field + " can not be blank")
		}
	}
	return nil
}

func validatePayoutCreateRequest(request APIRequest) error {
	if request == nil {
		return validationError("payout request can not be nil")
	}
	for _, field := range []string{"orderNo", "currency", "amount", "paymentMethod", "paymentMethodData"} {
		if strings.TrimSpace(fmt.Sprint(request[field])) == "" || fmt.Sprint(request[field]) == "<nil>" {
			return validationError(field + " can not be blank")
		}
	}
	return nil
}

func validateCustomerRequest(request APIRequest) error {
	if request == nil {
		return validationError("customer request can not be nil")
	}
	for _, field := range []string{"firstname", "lastname", "email", "country"} {
		if strings.TrimSpace(fmt.Sprint(request[field])) == "" || fmt.Sprint(request[field]) == "<nil>" {
			return validationError(field + " can not be blank")
		}
	}
	return nil
}
