package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	sdk "github.com/wikerx/payment-gateway-go-sdk"
)

type DemoAPI struct {
	Code        string
	Group       string
	Name        string
	Description string
	Action      string
	Method      string
	Path        string
	Fields      []DemoField
	Response    []ResponseField
}

type DemoField struct {
	Name        string
	Label       string
	Description string
	Required    bool
	Type        string
	Default     string
	Options     []string
}

type ResponseField struct {
	Name        string
	Description string
}

type PageData struct {
	MerchantNo  string
	Groups      map[string][]DemoAPI
	API         DemoAPI
	FormValues  map[string]string
	Request     string
	Response    string
	Error       string
	MethodsJSON template.JS
}

func main() {
	client, err := sdk.Create("config/merchant-config.properties")
	if err != nil {
		log.Fatal(err)
	}
	methodsJSON, _ := json.Marshal(sdk.PaymentMethodDataExamples())
	http.HandleFunc("/demo/apis", indexHandler(client, template.JS(methodsJSON)))
	http.HandleFunc("/demo/api/", apiHandler(client, template.JS(methodsJSON)))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/apis", http.StatusFound)
	})
	log.Println("Go SDK demo listening at http://127.0.0.1:58085/demo/apis")
	log.Fatal(http.ListenAndServe("127.0.0.1:58085", nil))
}

func indexHandler(client *sdk.Client, methodsJSON template.JS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, indexTemplate, PageData{
			MerchantNo:  client.Config().MerchantID,
			Groups:      groupCatalog(),
			MethodsJSON: methodsJSON,
		})
	}
}

func apiHandler(client *sdk.Client, methodsJSON template.JS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := strings.TrimPrefix(r.URL.Path, "/demo/api/")
		api, ok := apiCatalog()[code]
		if !ok {
			http.NotFound(w, r)
			return
		}
		data := PageData{MerchantNo: client.Config().MerchantID, API: api, MethodsJSON: methodsJSON}
		if r.Method == http.MethodPost {
			data.FormValues = formValuesFromPost(r, api, client.Config().MerchantID)
			request := buildRequest(r, api, client.Config().MerchantID)
			data.Request = sdk.ToPrettyJSON(request)
			result, err := invoke(r.Context(), client, api.Code, request)
			if err != nil {
				data.Error = err.Error()
			} else {
				data.Response = sdk.ToPrettyJSON(result)
			}
		} else {
			data.FormValues = defaultFormValues(api, client.Config().MerchantID)
			data.Request = sdk.ToPrettyJSON(requestFromValues(api, data.FormValues, client.Config().MerchantID))
		}
		render(w, apiTemplate, data)
	}
}

func invoke(ctx context.Context, client *sdk.Client, code string, request sdk.APIRequest) (*sdk.Result, error) {
	switch code {
	case "payin-checkout":
		return client.CreateCheckoutPayment(ctx, request)
	case "payin-direct":
		return client.CreateLocalPayment(ctx, request)
	case "payin-retrieve":
		return client.RetrievePayment(ctx, stringValue(request["tradeNo"]))
	case "refund-create":
		return client.CreateRefund(ctx, request)
	case "refund-retrieve":
		return client.RetrieveRefund(ctx, stringValue(request["refundNo"]))
	case "payout-create":
		return client.CreatePayout(ctx, request)
	case "payout-retrieve":
		return client.RetrievePayout(ctx, stringValue(request["tradeNo"]))
	case "payout-cancel":
		return client.CancelPayout(ctx, request)
	case "balance-retrieve":
		return client.RetrieveBalances(ctx, stringValue(request["currency"]))
	case "customer-create":
		return client.CreateCustomer(ctx, request)
	case "customer-retrieve":
		return client.RetrieveCustomer(ctx, stringValue(request["customerId"]))
	case "customer-update":
		return client.UpdateCustomer(ctx, stringValue(request["customerId"]), request)
	case "customer-delete":
		return client.DeleteCustomer(ctx, stringValue(request["customerId"]))
	case "customer-list":
		return client.ListCustomers(ctx)
	default:
		return nil, &sdk.SDKError{Kind: sdk.ErrorKindValidation, Msg: "unsupported demo api"}
	}
}

func buildRequest(r *http.Request, api DemoAPI, merchantNo string) sdk.APIRequest {
	_ = r.ParseForm()
	return requestFromValues(api, formValuesFromPost(r, api, merchantNo), merchantNo)
}

func requestFromValues(api DemoAPI, values map[string]string, merchantNo string) sdk.APIRequest {
	request := sdk.APIRequest{"merchantNo": merchantNo}
	for _, field := range api.Fields {
		if field.Name == "merchantNo" {
			continue
		}
		value := values[field.Name]
		if value == "" {
			continue
		}
		switch field.Type {
		case "json":
			var decoded any
			if err := json.Unmarshal([]byte(value), &decoded); err == nil {
				request[field.Name] = decoded
			}
		case "array":
			request[field.Name] = []string{value}
		default:
			request[field.Name] = value
		}
	}
	if mode := values["customerMode"]; mode == "customer" {
		delete(request, "customerId")
	} else if mode == "customerId" {
		delete(request, "customer")
	}
	delete(request, "customerMode")
	return request
}

func defaultFormValues(api DemoAPI, merchantNo string) map[string]string {
	values := map[string]string{"merchantNo": merchantNo}
	for _, field := range api.Fields {
		values[field.Name] = fieldDefault(field, merchantNo)
	}
	return values
}

func formValuesFromPost(r *http.Request, api DemoAPI, merchantNo string) map[string]string {
	_ = r.ParseForm()
	values := defaultFormValues(api, merchantNo)
	values["merchantNo"] = merchantNo
	for _, field := range api.Fields {
		if field.Name == "merchantNo" {
			continue
		}
		if _, exists := r.Form[field.Name]; exists {
			values[field.Name] = r.FormValue(field.Name)
		}
	}
	return values
}

func render(w http.ResponseWriter, source string, data PageData) {
	tpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"fieldValue": fieldValue,
		"selected": func(option string, fieldName string, values map[string]string, defaultValue string) string {
			value := values[fieldName]
			if value == "" {
				value = defaultValue
			}
			if option == value {
				return "selected"
			}
			return ""
		},
	}).Parse(source))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, data)
}

func fieldValue(field DemoField, values map[string]string, merchantNo string) string {
	if value, ok := values[field.Name]; ok {
		return value
	}
	return fieldDefault(field, merchantNo)
}

func fieldDefault(field DemoField, merchantNo string) string {
	switch field.Default {
	case "CUSTOMER":
		return sdk.ToPrettyJSON(sdk.CustomerExample())
	case "PAYMENT_METHOD_DATA_CASHAPP":
		return sdk.ToPrettyJSON(sdk.PaymentMethodDataExamples()[sdk.PaymentMethodCashApp])
	case "PAYMENT_METHOD_DATA_CARD":
		return sdk.ToPrettyJSON(sdk.PaymentMethodDataExamples()[sdk.PaymentMethodCard])
	case "AUTO_PAYIN_CHECKOUT":
		return sdk.GenerateOrderNo("PAYIN_CHECKOUT_")
	case "AUTO_PAYIN_DIRECT":
		return sdk.GenerateOrderNo("PAYIN_DIRECT_")
	case "AUTO_PAYOUT":
		return sdk.GenerateOrderNo("PAYOUT_")
	default:
		return field.Default
	}
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	text := fmt.Sprint(value)
	if text == "<nil>" {
		return ""
	}
	return strings.TrimSpace(text)
}
