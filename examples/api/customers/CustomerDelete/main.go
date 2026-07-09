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
	customerID := "cus_xxx"
	result, err := client.DeleteCustomer(context.Background(), customerID)
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
