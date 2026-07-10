package payloadcrypto

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	payoutCreatePath = "/pay-api/payout/trade/transfer"
	defaultConfigRel = "../../../config/merchant-config.properties"
)

// TestMerchantPayoutCreateFullFlow performs a real payout-create HTTP request
// to payment.gateway.base-url from merchant-config.properties.
//
// The complete merchant-side flow is:
//  1. read merchant number, gateway URL, livemode, JWT secret, platform request
//     public key, and merchant response private key from merchant-config.properties;
//  2. build payout business parameters;
//  3. encrypt the request body with EncryptPayload;
//  4. build OpenAPI headers with BuildOpenAPIHeaders;
//  5. send POST /pay-api/payout/trade/transfer to the configured gateway;
//  6. parse the gateway response envelope and decrypt response data with
//     DecryptPayload.
func TestMerchantPayoutCreateFullFlow(t *testing.T) {
	configPath := resolveMerchantConfigPath()
	t.Logf("配置文件路径: %s", configPath)

	merchantConfig := loadStandaloneMerchantConfig(t, configPath)
	t.Logf("真实请求地址: %s", strings.TrimRight(merchantConfig.BaseURL, "/")+payoutCreatePath)

	result := callPayoutCreate(t, merchantConfig, buildPayoutCreateRequest(t, merchantConfig))
	debugLog("真实网关代付响应明文", sanitizeForLog(result))
}

// buildPayoutCreateRequest builds demo business parameters for the payout
// create API. Merchant number is read from merchant-config.properties; the
// remaining fields are sample payout parameters merchants should replace with
// their real order/customer/payment data.
func buildPayoutCreateRequest(t *testing.T, merchantConfig standaloneMerchantConfig) map[string]any {
	t.Helper()
	orderNo, err := GenerateOrderNo("PAYOUT_")
	if err != nil {
		t.Fatal(err)
	}
	notifyURL := strings.TrimSpace(os.Getenv("PAYOUT_NOTIFY_URL"))
	if notifyURL == "" {
		notifyURL = "http://192.168.2.114:58084/webhook/payout"
	}
	t.Logf("代付回调地址 notifyUrl: %s", notifyURL)
	return map[string]any{
		"merchantNo":    merchantConfig.MerchantNo,
		"orderNo":       orderNo,
		"amount":        "9.99",
		"currency":      "USD",
		"paymentMethod": "CARD",
		"clientIp":      "47.125.221.223",
		"notifyUrl":     notifyURL,
		"website":       "https://merchant.example.com",
		"customer": map[string]any{
			"firstname": "Lily",
			"lastname":  "Brown",
			"email":     "lily_brown@example.com",
			"phone":     "13628173752",
			"country":   "US",
			"state":     "CA",
			"city":      "Los Angeles",
			"address":   "123 Main St, Apt 4B",
			"zipcode":   "90001",
		},
		"paymentMethodData": map[string]any{
			"number":     "4000056655665556",
			"expMonth":   "06",
			"expYear":    "2029",
			"cvc":        "123",
			"holderName": "Lily Brown",
			"email":      "lily_brown@example.com",
		},
	}
}

// callPayoutCreate performs the complete merchant-side request:
// plaintext JSON -> EncryptPayload -> BuildOpenAPIHeaders -> HTTP POST ->
// response envelope parsing -> DecryptPayload.
func callPayoutCreate(t *testing.T, merchantConfig standaloneMerchantConfig, payoutRequest map[string]any) map[string]any {
	t.Helper()
	plainRequestJSON := mustJSON(t, payoutRequest)
	debugLog("请求原始明文报文", sanitizeForLog(payoutRequest))

	compactData, err := EncryptPayload(plainRequestJSON, merchantConfig.PlatformRequestPublicKey)
	if err != nil {
		t.Fatal(err)
	}
	httpBody := mustJSON(t, map[string]any{
		"livemode": merchantConfig.Livemode,
		"data":     compactData,
	})
	debugLog("请求密文参数", map[string]any{
		"livemode": merchantConfig.Livemode,
		"data":     compactData,
	})

	headers, err := BuildOpenAPIHeaders(merchantConfig.MerchantNo, merchantConfig.APIPrivateKey, merchantConfig.Livemode, true)
	if err != nil {
		t.Fatal(err)
	}
	debugLog("请求Header参数", sanitizeHeadersForLog(headers))

	requestURL := strings.TrimRight(merchantConfig.BaseURL, "/") + payoutCreatePath
	t.Logf("最终请求地址: %s", requestURL)
	debugLog("API调用开始", map[string]any{
		"apiName":    "Payout Transfer Create",
		"method":     http.MethodPost,
		"url":        requestURL,
		"merchantNo": merchantConfig.MerchantNo,
		"livemode":   merchantConfig.Livemode,
	})

	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewBufferString(httpBody))
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("unexpected HTTP status: %d", response.StatusCode)
	}

	var gatewayResponse struct {
		Code     int    `json:"code"`
		Msg      string `json:"msg"`
		Livemode bool   `json:"livemode"`
		Data     string `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&gatewayResponse); err != nil {
		t.Fatal(err)
	}
	debugLog("响应原始密文参数", gatewayResponse)
	if gatewayResponse.Livemode != merchantConfig.Livemode {
		t.Fatalf("response livemode mismatch: %v", gatewayResponse.Livemode)
	}
	if gatewayResponse.Code != 0 {
		t.Fatalf("gateway business error: code=%d msg=%s", gatewayResponse.Code, gatewayResponse.Msg)
	}

	plainResponseJSON, err := DecryptPayload(gatewayResponse.Data, merchantConfig.MerchantResponsePrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	var payoutResponse map[string]any
	if err := json.Unmarshal([]byte(plainResponseJSON), &payoutResponse); err != nil {
		t.Fatal(err)
	}
	debugLog("响应原始明文参数", sanitizeForLog(payoutResponse))
	return payoutResponse
}

// standaloneMerchantConfig is the minimal config shape needed by the standalone
// example. It mirrors the relevant keys from merchant-config.properties.
type standaloneMerchantConfig struct {
	BaseURL                    string
	MerchantNo                 string
	APIPrivateKey              string
	Livemode                   bool
	PlatformRequestPublicKey   string
	MerchantResponsePrivateKey string
}

// resolveMerchantConfigPath returns the config file path. Merchants can set
// PAYMENT_GATEWAY_CONFIG to point to their own merchant-config.properties;
// otherwise the SDK sample config is used.
func resolveMerchantConfigPath() string {
	if path := strings.TrimSpace(os.Getenv("PAYMENT_GATEWAY_CONFIG")); path != "" {
		return path
	}
	return defaultConfigRel
}

// loadStandaloneMerchantConfig reads merchant-config.properties and resolves
// inline key text or key file paths into standaloneMerchantConfig.
func loadStandaloneMerchantConfig(t *testing.T, path string) standaloneMerchantConfig {
	t.Helper()
	props := loadPropertiesFile(t, path)
	baseDir := filepath.Dir(path)
	return standaloneMerchantConfig{
		BaseURL:                    requireConfigValue(t, props, "payment.gateway.base-url"),
		MerchantNo:                 requireConfigValue(t, props, "payment.gateway.merchant-no"),
		APIPrivateKey:              requireConfigValue(t, props, "payment.gateway.api-private-key"),
		Livemode:                   requireConfigValue(t, props, "payment.gateway.livemode") == "true",
		PlatformRequestPublicKey:   resolveConfigKey(t, baseDir, props, "payment.gateway.platform-request-public-key-path", "payment.gateway.platform-request-public-key"),
		MerchantResponsePrivateKey: resolveConfigKey(t, baseDir, props, "payment.gateway.merchant-response-private-key-path", "payment.gateway.merchant-response-private-key"),
	}
}

// loadPropertiesFile parses a Java-style properties file with key=value or
// key:value lines, ignoring blank lines and comments.
func loadPropertiesFile(t *testing.T, path string) map[string]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	props := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		idx := strings.IndexAny(line, "=:")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		props[key] = value
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return props
}

// requireConfigValue returns a required config value or fails the test with the
// exact missing key name.
func requireConfigValue(t *testing.T, props map[string]string, key string) string {
	t.Helper()
	value := strings.TrimSpace(props[key])
	if value == "" {
		t.Fatalf("missing required config: %s", key)
	}
	return value
}

// resolveConfigKey loads a key from either a file path property or an inline
// text property. This matches the two key modes supported by the full SDK.
func resolveConfigKey(t *testing.T, baseDir string, props map[string]string, pathKey string, inlineKey string) string {
	t.Helper()
	keyPath := strings.TrimSpace(props[pathKey])
	if keyPath != "" {
		if strings.HasPrefix(keyPath, "classpath:") {
			keyPath = filepath.Join(baseDir, "..", strings.TrimPrefix(keyPath, "classpath:"))
		} else if !filepath.IsAbs(keyPath) {
			keyPath = filepath.Join(baseDir, keyPath)
		}
		bytes, err := os.ReadFile(keyPath)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(bytes))
	}
	return requireConfigValue(t, props, inlineKey)
}

// mustJSON marshals a value to compact JSON and fails the test on error.
func mustJSON(t *testing.T, value any) string {
	t.Helper()
	bytes, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}

// debugLog prints indented JSON debug output similar to the full SDK logs.
func debugLog(title string, value any) {
	log.Printf("%s: %s", title, prettyJSON(value))
}

// prettyJSON formats values for readable test logs.
func prettyJSON(value any) string {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

// sanitizeHeadersForLog masks Authorization while preserving other headers for
// troubleshooting.
func sanitizeHeadersForLog(headers map[string]string) map[string]string {
	result := make(map[string]string, len(headers))
	for key, value := range headers {
		if strings.EqualFold(key, "Authorization") {
			result[key] = mask(value)
			continue
		}
		result[key] = value
	}
	return result
}

// sanitizeForLog recursively masks common payment, customer, and account fields
// before printing plaintext request or response data.
func sanitizeForLog(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, val := range typed {
			switch key {
			case "number", "cardNo", "cvc", "cvv", "accountNumber", "routingNumber", "phone", "email":
				result[key] = mask(stringValue(val))
			default:
				result[key] = sanitizeForLog(val)
			}
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, val := range typed {
			result[i] = sanitizeForLog(val)
		}
		return result
	default:
		return value
	}
}

// mask keeps a short prefix/suffix and hides the middle of sensitive values.
func mask(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}

// stringValue converts a log value to string without panicking.
func stringValue(value any) string {
	bytes, err := json.Marshal(value)
	if err == nil {
		return strings.Trim(string(bytes), `"`)
	}
	return ""
}
