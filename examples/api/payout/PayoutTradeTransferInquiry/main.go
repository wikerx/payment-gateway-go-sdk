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
	tradeNo := os.Getenv("PAYOUT_TRADE_NO")
	if tradeNo == "" {
		panic("set PAYOUT_TRADE_NO to a real payout tradeNo returned by payout create")
	}
	result, err := client.RetrievePayout(context.Background(), tradeNo)
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
