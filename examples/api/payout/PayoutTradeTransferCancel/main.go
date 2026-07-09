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
	result, err := client.CancelPayout(context.Background(), sdk.APIRequest{
		"merchantNo": client.Config().MerchantID,
		"tradeNo":    "payout_xxx",
		"orderNo":    "PAYOUT_xxx",
		"remark":     "Go SDK demo payout cancel",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
