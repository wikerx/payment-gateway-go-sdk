package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

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
	http.HandleFunc("/webhook/payout", func(w http.ResponseWriter, r *http.Request) {
		logIncomingCallback("Payout callback request", r)
		fmt.Println("Payout signature debug:")
		fmt.Println(sdk.ToPrettyJSON(verifier.DebugPayoutCallbackSignature(r)))
		payload, err := verifier.VerifyPayoutCallbackRequest(r)
		fmt.Printf("Payout signature verified: %v\n", err == nil)
		if err != nil {
			fmt.Printf("Payout signature verify error: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println("Payout callback payload:")
		fmt.Println(sdk.ToPrettyJSON(payload))
		_, _ = w.Write([]byte("success"))
	})
	port := os.Getenv("WEBHOOK_PORT")
	if port == "" {
		port = "58084"
	}
	fmt.Printf("Payout webhook listening on http://127.0.0.1:%s/webhook/payout\n", port)
	panic(http.ListenAndServe(":"+port, nil))
}

func logIncomingCallback(title string, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	fmt.Println(title + ":")
	fmt.Println(sdk.ToPrettyJSON(map[string]any{
		"method":      r.Method,
		"path":        r.URL.Path,
		"rawQuery":    r.URL.RawQuery,
		"queryParams": r.URL.Query(),
		"headers":     r.Header,
		"body":        string(body),
	}))
}
