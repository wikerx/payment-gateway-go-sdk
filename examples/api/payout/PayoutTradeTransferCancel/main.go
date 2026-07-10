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
	// Cancel must use a real tradeNo and its original merchant orderNo returned
	// by a successful payout-create request.
	tradeNo := os.Getenv("PAYOUT_TRADE_NO")
	orderNo := os.Getenv("PAYOUT_ORDER_NO")
	if tradeNo == "" || orderNo == "" {
		panic("set PAYOUT_TRADE_NO and PAYOUT_ORDER_NO from a real payout-create response")
	}
	result, err := client.CancelPayout(context.Background(), sdk.APIRequest{
		"merchantNo": client.Config().MerchantID,
		"tradeNo":    tradeNo,
		"orderNo":    orderNo,
		"remark":     "Go SDK demo payout cancel",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
