# Standalone Payload Crypto

这个目录提供一个可以直接复制到商户项目里的 OpenAPI 报文体加解密实现。

## 商户需要复制哪个文件

只需要复制：

```text
payload_crypto.go
```

它是单文件实现，只依赖 Go 标准库，不依赖本 SDK 里的 `crypto.go`、`keys.go`、`config.go`、`client.go` 等内部文件。

复制后可以把文件第一行包名：

```go
package payloadcrypto
```

改成商户自己项目里的包名。

## 函数

```go
compactData, err := EncryptPayload(
    `{"orderNo":"PAYIN_1","amount":"12.34","currency":"USD"}`,
    platformRequestPublicKeyText,
)
```

```go
plainJSON, err := DecryptPayload(
    compactData,
    merchantResponsePrivateKeyText,
)
```

## 密钥格式

| 参数 | 支持格式 |
|---|---|
| `platformPublicKeyText` | 平台请求公钥 PEM 文本或 X.509 DER Base64 文本 |
| `merchantResponsePrivateKeyText` | 商户响应私钥 PEM 文本、PKCS#8 DER Base64 文本或 PKCS#1 RSA 私钥 PEM 文本 |

## 边界

这个文件只负责 compact `data` 的加密和解密：

- 不生成 JWT
- 不发送 HTTP 请求
- 不读取 `merchant-config.properties`
- 不处理 `livemode`
- 不解析网关响应外壳 `code/msg/livemode`
