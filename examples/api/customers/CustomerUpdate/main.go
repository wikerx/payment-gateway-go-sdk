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
	customerID := os.Getenv("CUSTOMER_ID")
	if customerID == "" {
		panic("set CUSTOMER_ID to a real customerId returned by customer create")
	}
	request := sdk.CustomerExample()
	request["merchantNo"] = client.Config().MerchantID
	request["city"] = "San Francisco"
	result, err := client.UpdateCustomer(context.Background(), customerID, request)
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
