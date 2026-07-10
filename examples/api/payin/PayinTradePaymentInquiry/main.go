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
	tradeNo := os.Getenv("PAYMENT_TRADE_NO")
	if tradeNo == "" {
		panic("set PAYMENT_TRADE_NO to a real payin tradeNo returned by payment create")
	}
	result, err := client.RetrievePayment(context.Background(), tradeNo)
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
