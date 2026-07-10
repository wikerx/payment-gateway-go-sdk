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
	result, err := client.CreateLocalPayment(context.Background(), sdk.APIRequest{
		"merchantNo":        client.Config().MerchantID,
		"orderNo":           sdk.GenerateOrderNo("PAYIN_DIRECT_"),
		"payType":           sdk.PaymentTypeDirect,
		"currency":          "USD",
		"amount":            "12.34",
		"paymentMethod":     sdk.PaymentMethodCashApp,
		"paymentMethodData": sdk.PaymentMethodDataExamples()[sdk.PaymentMethodCashApp],
		"notifyUrl":         notifyURL,
		"clientIp":          "47.125.221.223",
		"website":           "http://127.0.0.1:5173",
		"metadata":          "metadata",
		"customer":          sdk.CustomerExample(),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sdk.ToPrettyJSON(result))
}
