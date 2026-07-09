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
	request := sdk.CustomerExample()
	request["merchantNo"] = client.Config().MerchantID
	request["email"] = "lily_brown_" + sdk.GenerateOrderNo("") + "@test.com"
	result, err := client.CreateCustomer(context.Background(), request)
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
