# OpenAPI 协议说明

## 请求认证

SDK 每次请求都会生成 `Authorization: Bearer <jwt>`。

JWT 使用 HS256，包含：

| Claim | 说明 |
|---|---|
| `iss` | 固定为 `merchant` |
| `aud` | 固定为 `gateway` |
| `jti` | 每次请求唯一，用于防重放 |
| `iat` | 签发时间 |
| `exp` | 过期时间，默认 180 秒内 |
| `merchantId` | 商户号 |
| `livemode` | 沙盒/生产标识 |

Go SDK 的 `jti` 由业务前缀、纳秒时间和 16 字节安全随机数组成，例如：

```text
PAYOUT_CREATE_1783591234567890000_7d8f...
```

`jti` 只用于网关鉴权层防重放，不承担业务幂等职责。商户订单号、交易号、退款号仍应放在加密业务请求体中。

## 请求加密

有请求体的接口使用：

```json
{
  "livemode": false,
  "data": "protectedHeader.encryptedAesKey.iv.cipherText.tag"
}
```

`data` 是 compact payload：

| 段 | 说明 |
|---|---|
| `protectedHeader` | Base64URL JSON，包含 `typ=PAYMENT-PAYLOAD`、`alg=RSA-OAEP-256`、`enc=A256GCM` |
| `encryptedAesKey` | 使用平台请求公钥 RSA-OAEP-SHA256 加密后的 AES-256 会话密钥 |
| `iv` | AES-GCM 12 字节随机 IV |
| `cipherText` | AES-256-GCM 密文 |
| `tag` | AES-GCM 认证标签 |

`protectedHeader` 会作为 AES-GCM AAD 参与认证。

Go SDK 提供两个单独的报文体加解密函数，方便商户独立验证协议：

```go
compactData, err := paymentgateway.EncryptPayload(plainJSON, platformRequestPublicKeyText)
plainJSON, err := paymentgateway.DecryptPayload(compactData, merchantResponsePrivateKeyText)
```

这两个函数只处理 compact `data`，不处理 JWT、HTTP Header、`livemode` 外壳或网关业务响应码。

如果商户不想导入完整 SDK，只想复制加解密代码，可以直接复制：

```text
examples/standalone/payloadcrypto/payload_crypto.go
```

该文件是单文件独立实现，只依赖 Go 标准库，内部已经包含密钥解析、RSA-OAEP-256、AES-256-GCM、Base64URL compact 拼装和解析逻辑。

## 响应解密

网关响应外壳：

```json
{
  "code": 0,
  "msg": "success",
  "livemode": false,
  "data": "protectedHeader.encryptedAesKey.iv.cipherText.tag"
}
```

SDK 使用商户响应私钥解密 `data`，并返回 `Result`。HTTP 2xx 不代表业务成功，商户仍需检查 `code` 和业务状态字段。

开启 `payment.gateway.debug-raw-log-enabled=true` 时，SDK 会在解密前打印 `响应原始密文参数`，并把响应 compact payload 拆分为：

| 字段 | 说明 |
|---|---|
| `protectedHeader` | Base64URL 编码的响应 protected header |
| `header` | 解码后的响应 header JSON |
| `encryptedAesKey` | RSA-OAEP-256 加密后的 AES 会话密钥 |
| `iv` | AES-GCM IV |
| `cipherText` | 响应密文 |
| `tag` | AES-GCM 认证标签 |

## 路径

| API | 方法 | 路径 |
|---|---|---|
| 创建代收 | POST | `/pay-api/trade/payment` |
| 查询代收 | GET | `/pay-api/trade/payment/{tradeNo}` |
| 创建退款 | POST | `/pay-api/trade/refund` |
| 查询退款 | GET | `/pay-api/trade/refund/{refundNo}` |
| 创建代付 | POST | `/pay-api/payout/trade/transfer` |
| 查询代付 | GET | `/pay-api/payout/trade/transfer/{tradeNo}` |
| 取消代付 | POST | `/pay-api/payout/trade/transfer-cancel` |
| 查询余额 | GET | `/pay-api/fund/accounts/get` |
| 创建客户 | POST | `/pay-api/mer/customers` |
| 查询客户 | GET | `/pay-api/mer/customers/{customerId}` |
| 更新客户 | PUT | `/pay-api/mer/customers/{customerId}` |
| 删除客户 | DELETE | `/pay-api/mer/customers/{customerId}` |
| 客户列表 | GET | `/pay-api/mer/customers` |
