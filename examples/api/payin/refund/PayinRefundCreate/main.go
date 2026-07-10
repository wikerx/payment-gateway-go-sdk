package main

import (
	"context"
	"fmt"
	"os"

	sdk "github.com/wikerx/payment-gateway-go-sdk"
)

func main() {
	client, err := sdk.Create("config/merchant-config.properties")
	if err != nil {
		panic(err)
	}
	// Refund must use a real payin tradeNo returned by a successful payment.
	tradeNo := os.Getenv("PAYMENT_TRADE_NO")
	if tradeNo == "" {
		panic("set PAYMENT_TRADE_NO to a real refundable payin tradeNo")
	}
	result, err := client.CreateRefund(context.Background(), sdk.APIRequest{
		"merchantNo":   client.Config().MerchantID,
		"tradeNo":      tradeNo,
		"currency":     "USD",
		"amount":       "12.34",
		"refundAmount": "1.00",
		"refundReason": "Go SDK demo refund request",
		"metadata":     "metadata",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
