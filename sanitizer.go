package paymentgateway

import (
	"fmt"
	"strings"
)

var sensitiveFieldNames = map[string]bool{
	"number": true, "cardNo": true, "cvc": true, "cvv": true,
	"accountNumber": true, "routingNumber": true, "phone": true,
	"email": true, "api_private_key": true, "apiPrivateKey": true,
	"merchantJwtSecret": true, "merchant_response_private_key": true,
	"merchantResponsePrivateKey": true, "platform_request_public_key": true,
	"platformPublicKey": true, "Authorization": true, "authorization": true,
}

func Sanitize(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, val := range typed {
			if sensitiveFieldNames[key] {
				result[key] = mask(fmt.Sprint(val))
			} else {
				result[key] = Sanitize(val)
			}
		}
		return result
	case map[string]string:
		result := make(map[string]string, len(typed))
		for key, val := range typed {
			if sensitiveFieldNames[key] {
				result[key] = mask(val)
			} else {
				result[key] = val
			}
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, val := range typed {
			result[i] = Sanitize(val)
		}
		return result
	default:
		return value
	}
}

func SanitizeHeaders(headers map[string][]string) map[string][]string {
	result := make(map[string][]string, len(headers))
	for key, values := range headers {
		if sensitiveFieldNames[key] {
			result[key] = []string{mask(strings.Join(values, ","))}
			continue
		}
		copied := make([]string, len(values))
		copy(copied, values)
		result[key] = copied
	}
	return result
}

func EncryptedDataSummary(data string) string {
	if len(data) <= 24 {
		return mask(data)
	}
	return data[:12] + "..." + data[len(data)-12:]
}

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
