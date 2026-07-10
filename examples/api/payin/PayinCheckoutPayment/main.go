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
	notifyURL := os.Getenv("PAYIN_NOTIFY_URL")
	if notifyURL == "" {
		notifyURL = "http://127.0.0.1:58083/webhook/payin"
	}
	result, err := client.CreateCheckoutPayment(context.Background(), sdk.APIRequest{
		"merchantNo":         client.Config().MerchantID,
		"orderNo":            sdk.GenerateOrderNo("PAYIN_CHECKOUT_"),
		"currency":           "USD",
		"amount":             "12.34",
		"returnUrl":          "https://manage.forgottenthrone.com/",
		"notifyUrl":          notifyURL,
		"clientIp":           "47.125.221.223",
		"website":            "https://manage.forgottenthrone.com/",
		"metadata":           "metadata",
		"paymentMethodTypes": []string{sdk.PaymentMethodCard},
		"customer":           sdk.CustomerExample(),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
