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
	notifyURL := os.Getenv("PAYOUT_NOTIFY_URL")
	if notifyURL == "" {
		notifyURL = "http://127.0.0.1:58084/webhook/payout"
	}
	result, err := client.CreatePayout(context.Background(), sdk.APIRequest{
		"merchantNo":        client.Config().MerchantID,
		"orderNo":           sdk.GenerateOrderNo("PAYOUT_"),
		"currency":          "USD",
		"amount":            "3.11",
		"paymentMethod":     sdk.PaymentMethodCard,
		"paymentMethodData": sdk.PaymentMethodDataExamples()[sdk.PaymentMethodCard],
		"notifyUrl":         notifyURL,
		"clientIp":          "47.125.221.223",
		"website":           "https://manage.forgottenthrone.com/",
		"metadata":          "metadata",
		"customer":          sdk.CustomerExample(),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
