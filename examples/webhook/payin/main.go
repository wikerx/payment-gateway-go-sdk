package main

import (
	"fmt"
	"io"
	"net/http"

	sdk "github.com/wikerx/payment-gateway-go-sdk"
)

func main() {
	config, err := sdk.LoadConfig("config/merchant-config.properties")
	if err != nil {
		panic(err)
	}
	verifier, err := sdk.NewWebhookVerifier(config)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/webhook/payin", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		payload, err := verifier.Verify(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println(sdk.ToPrettyJSON(payload))
		_, _ = w.Write([]byte("success"))
	})
	panic(http.ListenAndServe(":58083", nil))
}
