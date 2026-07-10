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
	result, err := client.DeleteCustomer(context.Background(), customerID)
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
