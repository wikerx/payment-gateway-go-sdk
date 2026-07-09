package paymentgateway

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BaseURL                    string
	MerchantID                 string
	MerchantJWTSecret          string
	Livemode                   bool
	PlatformPublicKey          string
	MerchantResponsePrivateKey string
	JWTTTLSeconds              int
	ConnectTimeout             time.Duration
	ReadTimeout                time.Duration
	DefaultVersion             string
	RawHTTPLogEnabled          bool
}

func (c *Config) Validate() error {
	c.BaseURL = strings.TrimRight(requireTrim(c.BaseURL), "/")
	c.MerchantID = requireTrim(c.MerchantID)
	c.MerchantJWTSecret = requireTrim(c.MerchantJWTSecret)
	c.PlatformPublicKey = requireTrim(c.PlatformPublicKey)
	c.MerchantResponsePrivateKey = requireTrim(c.MerchantResponsePrivateKey)
	if c.BaseURL == "" {
		return configError("payment.gateway.base-url can not be blank", nil)
	}
	if c.MerchantID == "" {
		return configError("payment.gateway.merchant-no can not be blank", nil)
	}
	if c.MerchantJWTSecret == "" {
		return configError("payment.gateway.api-private-key can not be blank", nil)
	}
	if c.PlatformPublicKey == "" {
		return configError("payment.gateway.platform-request-public-key can not be blank", nil)
	}
	if c.MerchantResponsePrivateKey == "" {
		return configError("payment.gateway.merchant-response-private-key can not be blank", nil)
	}
	if c.JWTTTLSeconds == 0 {
		c.JWTTTLSeconds = JWTTTLSeconds
	}
	if c.JWTTTLSeconds <= 0 || c.JWTTTLSeconds > JWTTTLSeconds {
		return configError("jwt ttlSeconds must be between 1 and 180", nil)
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = HTTPConnectTimeoutMS * time.Millisecond
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = HTTPReadTimeoutMS * time.Millisecond
	}
	if c.DefaultVersion == "" {
		c.DefaultVersion = DefaultVersion
	}
	return nil
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = filepath.Join("config", ConfigFileName)
	}
	props, err := loadProperties(path)
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Dir(path)
	livemode, err := parseRequiredBool(props, "payment.gateway.livemode")
	if err != nil {
		return nil, err
	}
	rawLog, err := parseOptionalBool(props, "payment.gateway.debug-raw-log-enabled", false)
	if err != nil {
		return nil, err
	}
	platformKey, err := resolveKey(baseDir, props["payment.gateway.platform-request-public-key-path"], props["payment.gateway.platform-request-public-key"], "payment.gateway.platform-request-public-key")
	if err != nil {
		return nil, err
	}
	responseKey, err := resolveKey(baseDir, props["payment.gateway.merchant-response-private-key-path"], props["payment.gateway.merchant-response-private-key"], "payment.gateway.merchant-response-private-key")
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		BaseURL:                    props["payment.gateway.base-url"],
		MerchantID:                 props["payment.gateway.merchant-no"],
		MerchantJWTSecret:          props["payment.gateway.api-private-key"],
		Livemode:                   livemode,
		PlatformPublicKey:          platformKey,
		MerchantResponsePrivateKey: responseKey,
		JWTTTLSeconds:              JWTTTLSeconds,
		ConnectTimeout:             HTTPConnectTimeoutMS * time.Millisecond,
		ReadTimeout:                HTTPReadTimeoutMS * time.Millisecond,
		DefaultVersion:             DefaultVersion,
		RawHTTPLogEnabled:          rawLog,
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadProperties(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, configError("failed to load merchant config file: "+path, err)
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
		return nil, configError("failed to read merchant config file: "+path, err)
	}
	return props, nil
}

func parseRequiredBool(props map[string]string, key string) (bool, error) {
	value := requireTrim(props[key])
	if value == "" {
		return false, configError("missing required merchant config: "+key, nil)
	}
	return parseBool(value, key)
}

func parseOptionalBool(props map[string]string, key string, fallback bool) (bool, error) {
	value := requireTrim(props[key])
	if value == "" {
		return fallback, nil
	}
	return parseBool(value, key)
}

func parseBool(value, key string) (bool, error) {
	parsed, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(value)))
	if err != nil {
		return false, configError("merchant config "+key+" must be true or false", err)
	}
	return parsed, nil
}

func resolveKey(baseDir, keyPath, inlineKey, name string) (string, error) {
	keyPath = requireTrim(keyPath)
	if keyPath != "" {
		if strings.HasPrefix(keyPath, "classpath:") {
			keyPath = strings.TrimPrefix(keyPath, "classpath:")
			keyPath = filepath.Join(baseDir, "..", keyPath)
		} else if !filepath.IsAbs(keyPath) {
			keyPath = filepath.Join(baseDir, keyPath)
		}
		bytes, err := os.ReadFile(keyPath)
		if err != nil {
			return "", configError("failed to read key file for "+name, err)
		}
		return strings.TrimSpace(string(bytes)), nil
	}
	inlineKey = requireTrim(inlineKey)
	if inlineKey == "" {
		return "", configError("missing required merchant config: "+name, nil)
	}
	return inlineKey, nil
}

func requireTrim(value string) string {
	return strings.TrimSpace(value)
}
