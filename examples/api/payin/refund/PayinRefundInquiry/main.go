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
	refundNo := os.Getenv("REFUND_NO")
	if refundNo == "" {
		panic("set REFUND_NO to a real refundNo returned by refund create")
	}
	result, err := client.RetrieveRefund(context.Background(), refundNo)
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
