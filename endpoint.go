package paymentgateway

import "fmt"

type Endpoint struct {
	// Name is a human-readable API name used in SDK logs.
	Name string
	// Method is the HTTP method used by the endpoint.
	Method string
	// Path is the gateway path. Query/path parameters are formatted separately.
	Path string
}

// Format injects path arguments into endpoints such as /resource/%s.
func (e Endpoint) Format(args ...any) string {
	if len(args) == 0 {
		return e.Path
	}
	return fmt.Sprintf(e.Path, args...)
}

// OpenAPI endpoint metadata used by the client and demo catalog.
var (
	EndpointPaymentCreate    = Endpoint{"Payment Create", "POST", PaymentCreatePath}
	EndpointPaymentRetrieve  = Endpoint{"Payment Retrieve", "GET", PaymentRetrievePath}
	EndpointRefundCreate     = Endpoint{"Refund Create", "POST", RefundCreatePath}
	EndpointRefundRetrieve   = Endpoint{"Refund Retrieve", "GET", RefundRetrievePath}
	EndpointPayoutCreate     = Endpoint{"Payout Transfer Create", "POST", PayoutCreatePath}
	EndpointPayoutRetrieve   = Endpoint{"Payout Transfer Retrieve", "GET", PayoutRetrievePath}
	EndpointPayoutCancel     = Endpoint{"Payout Transfer Cancel", "POST", PayoutCancelPath}
	EndpointBalanceInquiry   = Endpoint{"Fund Accounts Balance Inquiry", "GET", BalanceRetrievePath}
	EndpointCustomerCreate   = Endpoint{"Customer Create", "POST", CustomerCreatePath}
	EndpointCustomerRetrieve = Endpoint{"Customer Retrieve", "GET", CustomerRetrievePath}
	EndpointCustomerUpdate   = Endpoint{"Customer Update", "PUT", CustomerUpdatePath}
	EndpointCustomerDelete   = Endpoint{"Customer Delete", "DELETE", CustomerDeletePath}
	EndpointCustomerList     = Endpoint{"Customer List", "GET", CustomerListPath}
)
