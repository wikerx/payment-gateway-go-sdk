package main

import (
	"context"
	"fmt"

	sdk "github.com/wikerx/payment-gateway-go-sdk"
)

func main() {
	client, err := sdk.Create("config/merchant-config.properties")
	if err != nil {
		panic(err)
	}
	result, err := client.CreateRefund(context.Background(), sdk.APIRequest{
		"merchantNo":   client.Config().MerchantID,
		"tradeNo":      "pay_xxx",
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
