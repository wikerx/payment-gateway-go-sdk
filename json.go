package paymentgateway

import (
	"bytes"
	"encoding/json"
)

func toJSON(value any) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func toPrettyJSON(value any) string {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

// ToPrettyJSON formats a value as indented JSON for examples and demo output.
func ToPrettyJSON(value any) string {
	return toPrettyJSON(value)
}

func fromJSON(data string, target any) error {
	decoder := json.NewDecoder(bytes.NewBufferString(data))
	decoder.UseNumber()
	return decoder.Decode(target)
}

func fromJSONBytes(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(target)
}
