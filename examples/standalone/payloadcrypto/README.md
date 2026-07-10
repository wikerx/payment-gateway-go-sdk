# Standalone Payload Crypto

这个目录提供一个可以直接复制到商户项目里的 OpenAPI 报文体加解密和请求 Header 封装实现。

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

### 报文体加密

```go
compactData, err := EncryptPayload(
    `{"message":"payload crypto only","amount":"12.34","currency":"USD"}`,
    platformRequestPublicKeyText,
)
```

请求体发送时包装为：

```json
{
  "livemode": false,
  "data": "<compactData>"
}
```

### 响应报文体解密

```go
plainJSON, err := DecryptPayload(
    compactData,
    merchantResponsePrivateKeyText,
)
```

### 请求 Header 封装

```go
merchantConfig := loadStandaloneMerchantConfig("config/merchant-config.properties")
headers, err := BuildOpenAPIHeaders(
    merchantConfig.MerchantNo,
    merchantConfig.APIPrivateKey,
    merchantConfig.Livemode,
    true,
)
```

返回的 Header 包含：

| Header | 说明 |
|---|---|
| `Authorization` | `Bearer <merchant jwt>` |
| `Accept` | `application/json` |
| `User-Agent` | standalone helper 标识 |
| `X-Request-Id` | 每次请求自动生成 |
| `Content-Type` | `withBody=true` 时返回 `application/json; charset=UTF-8` |

如果商户想自己组 Header，也可以只生成 JWT：

```go
token, err := SignMerchantJWT(
    merchantConfig.MerchantNo,
    merchantConfig.APIPrivateKey,
    merchantConfig.Livemode,
    "JWT_202607100001",
    time.Now().UTC(),
    180,
)
```

`jwtID` 会写入 JWT 的 `jti`，每次请求必须唯一，不能重复使用，否则网关可能按防重放规则拒绝请求。

### 商户订单号生成

```go
orderNo, err := GenerateOrderNo("PAYOUT_")
```

创建代收、代付等会落库的请求，`orderNo` 必须每次新生成。不要复用示例订单号，否则网关会返回类似：

```json
{
  "code": 1040000009,
  "msg": "Site order number cannot be repeated"
}
```

## 完整代付流程测试

商户可以参考：

```text
payout_create_flow_test.go
```

它演示了完整代付请求链路：

1. 从 `merchant-config.properties` 读取商户号、网关地址、JWT 密钥、平台请求公钥、商户响应私钥、`livemode`
2. 自动生成新的 `orderNo`
3. 封装代付业务参数
4. 使用平台请求公钥加密请求 `data`
5. 生成 `Authorization`、`Content-Type`、`X-Request-Id` 等请求 Header
6. 发起 `POST /pay-api/payout/trade/transfer`
7. 解析网关响应外壳 `code/msg/livemode/data`
8. 使用商户响应私钥解密响应 `data`

运行：

```bash
go test ./examples/standalone/payloadcrypto -run TestMerchantPayoutCreateFullFlow -v
```

该测试会真实请求 `merchant-config.properties` 中的 `payment.gateway.base-url`：

```text
POST /pay-api/payout/trade/transfer
```

运行时会打印：

- API 调用开始
- 请求原始明文报文
- 请求密文参数
- 请求 Header 参数
- 响应原始密文参数
- 响应原始明文参数

如配置文件不在默认位置，可以指定：

```bash
PAYMENT_GATEWAY_CONFIG=/absolute/path/merchant-config.properties go test ./examples/standalone/payloadcrypto -run TestMerchantPayoutCreateFullFlow -v
```

请只在沙盒商户和确认参数无误时运行该测试。这个测试不是 mock，不会自己造响应数据；它会真实请求网关并创建测试交易。

## 密钥格式

| 参数 | 支持格式 |
|---|---|
| `platformPublicKeyText` | 平台请求公钥 PEM 文本或 X.509 DER Base64 文本 |
| `merchantResponsePrivateKeyText` | 商户响应私钥 PEM 文本、PKCS#8 DER Base64 文本或 PKCS#1 RSA 私钥 PEM 文本 |

## 边界

这个文件只负责：

- compact `data` 的加密和解密
- 商户 JWT 签名
- OpenAPI 请求 Header 组装
- 生成不会重复使用的示例商户订单号

它不负责：

- 发送 HTTP 请求
- 读取 `merchant-config.properties`
- 解析网关响应外壳 `code/msg/livemode`
