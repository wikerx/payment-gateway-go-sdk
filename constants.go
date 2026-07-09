package paymentgateway

const (
	// ConfigFileName is the default merchant properties file name.
	ConfigFileName = "merchant-config.properties"

	// DefaultVersion is reserved for versioned API routing.
	DefaultVersion = "v1"
	// JWTTTLSeconds is the maximum Authorization JWT lifetime accepted by the SDK.
	JWTTTLSeconds = 180
	// HTTPConnectTimeoutMS is the default connect timeout in milliseconds.
	HTTPConnectTimeoutMS = 3000
	// HTTPReadTimeoutMS is the default total request timeout in milliseconds.
	HTTPReadTimeoutMS = 10000
	// HTTPStatusSuccessMin is the inclusive lower bound for successful HTTP statuses.
	HTTPStatusSuccessMin = 200
	// HTTPStatusSuccessMax is the exclusive upper bound for successful HTTP statuses.
	HTTPStatusSuccessMax = 300
	// ContentType is used for encrypted JSON request envelopes.
	ContentType = "application/json; charset=UTF-8"
	// Accept asks the gateway for JSON response envelopes.
	Accept = "application/json"
	// ResponseCodeSuccess is the gateway business success code.
	ResponseCodeSuccess = 0
	// AuthorizationPrefix prefixes the bearer JWT in the Authorization header.
	AuthorizationPrefix = "Bearer "
	// HeaderAuthorization is the HTTP Authorization header name.
	HeaderAuthorization  = "Authorization"
	HeaderContentType    = "Content-Type"
	HeaderAccept         = "Accept"
	HeaderUserAgent      = "User-Agent"
	HeaderRequestID      = "X-Request-Id"
	PayloadHeaderType    = "typ"
	PayloadHeaderAlg     = "alg"
	PayloadHeaderEnc     = "enc"
	JWTHeaderType        = "typ"
	JWTType              = "JWT"
	SDKName              = "payment-gateway-go-sdk"
	SDKVersion           = "0.1.0"
	UserAgent            = SDKName + "/" + SDKVersion + " go"
	PayloadType          = "PAYMENT-PAYLOAD"
	KeyEncryptionAlg     = "RSA-OAEP-256"
	ContentEncryptionAlg = "A256GCM"
	PaymentCreatePath    = "/pay-api/trade/payment"
	PaymentRetrievePath  = "/pay-api/trade/payment/%s"
	RefundCreatePath     = "/pay-api/trade/refund"
	RefundRetrievePath   = "/pay-api/trade/refund/%s"
	PayoutCreatePath     = "/pay-api/payout/trade/transfer"
	PayoutRetrievePath   = "/pay-api/payout/trade/transfer/%s"
	PayoutCancelPath     = "/pay-api/payout/trade/transfer-cancel"
	BalanceRetrievePath  = "/pay-api/fund/accounts/get"
	CustomerCreatePath   = "/pay-api/mer/customers"
	CustomerRetrievePath = "/pay-api/mer/customers/%s"
	CustomerUpdatePath   = "/pay-api/mer/customers/%s"
	CustomerDeletePath   = "/pay-api/mer/customers/%s"
	CustomerListPath     = "/pay-api/mer/customers"
)
