package paymentgateway

const (
	ConfigFileName = "merchant-config.properties"

	DefaultVersion       = "v1"
	JWTTTLSeconds        = 180
	HTTPConnectTimeoutMS = 3000
	HTTPReadTimeoutMS    = 10000
	HTTPStatusSuccessMin = 200
	HTTPStatusSuccessMax = 300
	ContentType          = "application/json; charset=UTF-8"
	Accept               = "application/json"
	ResponseCodeSuccess  = 0
	AuthorizationPrefix  = "Bearer "
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
