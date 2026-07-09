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

func Create(configPath string, options ...ClientOption) (*Client, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return NewClient(cfg, options...)
}

type ClientOption func(*Client)

func WithHTTPTransport(transport HTTPTransport) ClientOption {
	return func(client *Client) {
		if transport != nil {
			client.transport = transport
		}
	}
}

func (c *Client) Config() *Config {
	return c.config
}

func (c *Client) CreateCheckoutPayment(ctx context.Context, request APIRequest) (*Result, error) {
	return c.createPayment(ctx, request)
}

func (c *Client) CreateLocalPayment(ctx context.Context, request APIRequest) (*Result, error) {
	return c.createPayment(ctx, request)
}

func (c *Client) CreateCardPayment(ctx context.Context, request APIRequest) (*Result, error) {
	if request == nil {
		return nil, validationError("payment request can not be nil")
	}
	if strings.TrimSpace(fmt.Sprint(request["paymentMethod"])) == "" {
		request["paymentMethod"] = PaymentMethodCard
	}
	return c.createPayment(ctx, request)
}

func (c *Client) RetrievePayment(ctx context.Context, tradeNo string) (*Result, error) {
	value, err := requireText(tradeNo, "tradeNo")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointPaymentRetrieve, uniqueJWTID("PAYMENT_QUERY_"), url.PathEscape(value))
}

func (c *Client) CreateRefund(ctx context.Context, request APIRequest) (*Result, error) {
	if err := validateRefundCreateRequest(request); err != nil {
		return nil, err
	}
	return c.postEncrypted(ctx, EndpointRefundCreate, request, uniqueJWTID("REFUND_CREATE_"))
}

func (c *Client) RetrieveRefund(ctx context.Context, refundNo string) (*Result, error) {
	value, err := requireText(refundNo, "refundNo")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointRefundRetrieve, uniqueJWTID("REFUND_QUERY_"), url.PathEscape(value))
}

func (c *Client) CreatePayout(ctx context.Context, request APIRequest) (*Result, error) {
	if err := validatePayoutCreateRequest(request); err != nil {
		return nil, err
	}
	return c.postEncrypted(ctx, EndpointPayoutCreate, request, uniqueJWTID("PAYOUT_CREATE_"))
}

func (c *Client) RetrievePayout(ctx context.Context, tradeNo string) (*Result, error) {
	value, err := requireText(tradeNo, "tradeNo")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointPayoutRetrieve, uniqueJWTID("PAYOUT_QUERY_"), url.PathEscape(value))
}

func (c *Client) CancelPayout(ctx context.Context, request APIRequest) (*Result, error) {
	if request == nil {
		return nil, validationError("request can not be nil")
	}
	return c.postEncrypted(ctx, EndpointPayoutCancel, request, uniqueJWTID("PAYOUT_CANCEL_"))
}

func (c *Client) RetrieveBalances(ctx context.Context, currency string) (*Result, error) {
	path := EndpointBalanceInquiry.Path
	if strings.TrimSpace(currency) != "" {
		path += "?currency=" + url.QueryEscape(strings.TrimSpace(currency))
	}
	return c.execute(ctx, EndpointBalanceInquiry, path, nil, nil, uniqueJWTID("BALANCE_QUERY_"))
}

func (c *Client) CreateCustomer(ctx context.Context, request APIRequest) (*Result, error) {
	if err := validateCustomerRequest(request); err != nil {
		return nil, err
	}
	return c.postEncrypted(ctx, EndpointCustomerCreate, request, uniqueJWTID("CUSTOMER_CREATE_"))
}

func (c *Client) RetrieveCustomer(ctx context.Context, customerID string) (*Result, error) {
	value, err := requireText(customerID, "customerId")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointCustomerRetrieve, uniqueJWTID("CUSTOMER_QUERY_"), url.PathEscape(value))
}

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

func (c *Client) DeleteCustomer(ctx context.Context, customerID string) (*Result, error) {
	value, err := requireText(customerID, "customerId")
	if err != nil {
		return nil, err
	}
	return c.getSecured(ctx, EndpointCustomerDelete, uniqueJWTID("CUSTOMER_DELETE_"), url.PathEscape(value))
}

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
