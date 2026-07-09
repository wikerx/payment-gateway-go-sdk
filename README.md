# Payment Gateway Go SDK

商户服务端 Go SDK，用于对接 Payment Gateway OpenAPI。SDK 会生成 `Authorization: Bearer <jwt>`，对写请求封装 `livemode + data` 加密外壳，并自动解密网关响应 `data`。

> 本 SDK 只能用于商户服务端。不要放在浏览器、移动端 App、桌面客户端或任何会暴露 API 私钥、RSA 私钥、卡号、CVC 的环境。

## 环境要求

- Go 1.22+
- 标准库 crypto/http/json 即可，无第三方运行时依赖

## 安装

```bash
go get github.com/wikerx/payment-gateway-go-sdk
```

本地联调时也可以在商户项目 `go.mod` 中使用：

```text
replace github.com/wikerx/payment-gateway-go-sdk => /Users/scott/Documents/code/idea_success/Zorpay/payment-gateway-go-sdk
```

## 配置

默认示例配置：

```text
config/merchant-config.properties
```

当前示例配置默认使用文本密钥模式，商户复制配置文件后可以直接联调测试商户；`keys/*.pem` 仍保留给 PEM 文件模式使用。

关键配置项：

| 配置项 | 必填 | 说明 |
|---|---:|---|
| `payment.gateway.base-url` | 是 | 网关地址，例如 `http://localhost:58060` |
| `payment.gateway.merchant-no` | 是 | 商户号 |
| `payment.gateway.livemode` | 是 | `false` 沙盒，`true` 生产 |
| `payment.gateway.api-private-key` | 是 | JWT HS256 签名密钥 |
| `payment.gateway.platform-request-public-key-path` | 二选一 | 平台请求公钥 PEM 路径 |
| `payment.gateway.merchant-response-private-key-path` | 二选一 | 商户响应私钥 PEM 路径 |
| `payment.gateway.platform-request-public-key` | 二选一 | 平台请求公钥文本 |
| `payment.gateway.merchant-response-private-key` | 二选一 | 商户响应私钥文本 |
| `payment.gateway.debug-raw-log-enabled` | 否 | 是否输出沙盒调试日志 |

### 文本密钥模式

默认启用：

```properties
payment.gateway.platform-request-public-key=<平台请求公钥 DER Base64 或 PEM 文本>
payment.gateway.merchant-response-private-key=<商户响应私钥 DER Base64 或 PEM 文本>
```

Go SDK 会自动解析 PEM 文本或 DER Base64 文本，再用于请求加密和响应解密。

### PEM 文件模式

如需使用 PEM 文件，请注释文本密钥配置，并打开：

```properties
payment.gateway.platform-request-public-key-path=../keys/2606177036_PLATFORM_REQUEST_PUBLIC_KEY.pem
payment.gateway.merchant-response-private-key-path=../keys/2606177036_MERCHANT_RESPONSE_PRIVATE_KEY.pem
```

路径支持绝对路径、相对配置文件目录的路径，也兼容 `classpath:` 前缀。

### 调试日志

`payment.gateway.debug-raw-log-enabled=true` 时会打印：

| 日志名称 | 内容 |
|---|---|
| `API调用开始` | API 名称、HTTP 方法、路径、商户号、`requestId`、JWT `jwtId` |
| `请求原始明文报文` | 脱敏后的业务请求参数 |
| `请求密文参数` | `livemode` 和请求 compact `data` |
| `响应原始密文参数` | HTTP 状态码、响应 Header、网关原始响应外壳 `code/msg/livemode/data` |
| `响应参数拆分` | 响应 compact payload 的 `protectedHeader/header/encryptedAesKey/iv/cipherText/tag` |
| `响应原始明文参数` | 解密后并脱敏的业务响应数据 |

生产环境建议关闭该开关。

## 创建客户端

```go
client, err := paymentgateway.Create("config/merchant-config.properties")
if err != nil {
    panic(err)
}
```

## 代收示例

```go
result, err := client.CreateLocalPayment(context.Background(), paymentgateway.APIRequest{
    "merchantNo": client.Config().MerchantID,
    "orderNo": paymentgateway.GenerateOrderNo("PAYIN_DIRECT_"),
    "payType": paymentgateway.PaymentTypeDirect,
    "currency": "USD",
    "amount": "12.34",
    "paymentMethod": paymentgateway.PaymentMethodCashApp,
    "paymentMethodData": paymentgateway.PaymentMethodDataExamples()[paymentgateway.PaymentMethodCashApp],
    "customer": paymentgateway.CustomerExample(),
})
```

## 已覆盖 API

| 分组 | API |
|---|---|
| 客户 | 创建客户、检索客户、更新客户、删除客户、列出客户 |
| 代收 | 创建收银台代收、创建直连代收、创建卡直连代收、查询代收 |
| 退款申请 | 创建退款、查询退款 |
| 代付 | 发起代付、查询代付、取消代付 |
| 余额查询 | 查询资金账户余额，可按币种过滤 |
| Webhook | 代收/代付回调密文解密与 livemode 校验 |

## 运行示例

```bash
go run ./examples/api/inquiry/balance/FundAccountsBalanceInquiry
go run ./examples/api/payin/PayinCheckoutPayment
go run ./examples/api/payin/PayinDirectPayment
go run ./examples/api/payout/PayoutTradeTransfer
go run ./examples/api/customers/CustomerCreate
```

查询、退款、取消示例中的 `pay_xxx`、`charge_xxx`、`payout_xxx`、`cus_xxx` 需要替换为真实测试接口返回值。

## 页面联调控制台

```bash
go run ./examples/demo
```

访问：

```text
http://127.0.0.1:58085/demo/apis
```

控制台按客户、代收、退款申请、代付、余额查询分组展示接口。创建代收、创建代付支持币种和支付方式下拉；直连代收和代付会根据 `paymentMethod` 自动替换 `paymentMethodData` 示例参数；代收支持 `customerId` 与 `customer` 二选一。

发起代付页面已包含 `notifyUrl`、`clientIp`、`website`、`metadata` 等网关校验字段，默认 `clientIp=47.125.221.223`。

## Webhook

```go
config, _ := paymentgateway.LoadConfig("config/merchant-config.properties")
verifier, _ := paymentgateway.NewWebhookVerifier(config)
payload, err := verifier.Verify(rawBody)
```

`Verify` 会解析回调外壳、校验 `livemode`，并使用商户响应私钥解密 `data`。

## 单独使用报文加解密

如果商户只想验证 OpenAPI 报文体加解密，不想使用完整 Client，可以直接使用两个纯函数：

```go
compactData, err := paymentgateway.EncryptPayload(
    `{"orderNo":"PAYIN_1","amount":"12.34","currency":"USD"}`,
    platformRequestPublicKeyText,
)
```

```go
plainJSON, err := paymentgateway.DecryptPayload(
    compactData,
    merchantResponsePrivateKeyText,
)
```

函数说明：

| 函数 | 作用 | 密钥格式 |
|---|---|---|
| `EncryptPayload(plainJSON, platformPublicKeyText)` | 把明文 JSON 加密成 `protectedHeader.encryptedAesKey.iv.cipherText.tag` | 平台请求公钥 PEM 文本或 X.509 DER Base64 |
| `DecryptPayload(compactData, merchantResponsePrivateKeyText)` | 把 compact `data` 解密成明文 JSON | 商户响应私钥 PEM 文本或 PKCS#8 DER Base64 |

这两个函数只处理报文体 `data`，不生成 JWT、不读取配置文件、不发送 HTTP 请求，也不解析网关响应外壳 `code/msg/livemode`。

## 验证

```bash
go test ./...
go test ./... -run TestPayloadCryptoRoundTrip
go test ./... -run TestStandalonePayloadCrypto
```

生产环境建议关闭 `payment.gateway.debug-raw-log-enabled`。
