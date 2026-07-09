package main

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPayoutCreateCatalogIncludesGatewayRequiredFields(t *testing.T) {
	api := apiCatalog()["payout-create"]
	fields := map[string]bool{}
	for _, field := range api.Fields {
		fields[field.Name] = true
	}
	for _, name := range []string{"clientIp", "notifyUrl", "website", "metadata"} {
		if !fields[name] {
			t.Fatalf("payout-create demo field %s is missing", name)
		}
	}
}

func TestPostedPaymentMethodIsPreservedForPayinDirect(t *testing.T) {
	api := apiCatalog()["payin-direct"]
	form := url.Values{}
	form.Set("paymentMethod", "CARD")
	form.Set("paymentMethodData", `{"number":"4000056655665556"}`)
	form.Set("orderNo", "PAYIN_1")
	form.Set("payType", "1")
	form.Set("currency", "USD")
	form.Set("amount", "12.34")
	request := httptest.NewRequest("POST", "/demo/api/payin-direct", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	values := formValuesFromPost(request, api, "2606177036")
	if values["paymentMethod"] != "CARD" {
		t.Fatalf("payment method should be preserved, got %s", values["paymentMethod"])
	}
	payload := requestFromValues(api, values, "2606177036")
	if payload["paymentMethod"] != "CARD" {
		t.Fatalf("request payment method should be CARD, got %#v", payload["paymentMethod"])
	}
}

func TestPostedPaymentMethodIsPreservedForPayoutCreate(t *testing.T) {
	api := apiCatalog()["payout-create"]
	form := url.Values{}
	form.Set("paymentMethod", "UPI")
	form.Set("paymentMethodData", `{"bankName":"test","cardNo":"6200000000000005"}`)
	form.Set("orderNo", "PAYOUT_1")
	form.Set("currency", "USD")
	form.Set("amount", "3.11")
	form.Set("clientIp", "47.125.221.223")
	request := httptest.NewRequest("POST", "/demo/api/payout-create", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	values := formValuesFromPost(request, api, "2606177036")
	if values["paymentMethod"] != "UPI" {
		t.Fatalf("payment method should be preserved, got %s", values["paymentMethod"])
	}
	payload := requestFromValues(api, values, "2606177036")
	if payload["paymentMethod"] != "UPI" {
		t.Fatalf("request payment method should be UPI, got %#v", payload["paymentMethod"])
	}
}
