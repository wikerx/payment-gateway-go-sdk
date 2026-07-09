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
	result, err := client.RetrieveRefund(context.Background(), "charge_xxx")
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
