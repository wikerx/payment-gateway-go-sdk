package main

func groupCatalog() map[string][]DemoAPI {
	groups := map[string][]DemoAPI{}
	for _, api := range apiCatalog() {
		groups[api.Group] = append(groups[api.Group], api)
	}
	return groups
}

func apiCatalog() map[string]DemoAPI {
	currency := []string{"USD", "EUR", "BRL", "MXN", "INR", "PHP"}
	methods := []string{"CARD", "CASHAPP", "PAY_PAL", "ACH_DEBIT", "UPI"}
	response := []ResponseField{
		{"code", "网关业务响应码，0 通常表示请求受理成功。"},
		{"msg", "网关业务响应描述。"},
		{"livemode", "响应环境标识，应与本地配置一致。"},
		{"data", "接口业务响应数据，不同 API 返回字段不同。"},
	}
	return map[string]DemoAPI{
		"customer-create": {"customer-create", "客户", "创建客户", "创建网关客户资料。", "创建客户", "POST", "/pay-api/mer/customers", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("firstname", "名", "客户名字。", true, "Lily"),
			text("lastname", "姓", "客户姓氏。", true, "Brown"),
			text("email", "邮箱", "客户邮箱。", true, "lily_brown_1782457030419@test.com"),
			text("phone", "手机号", "客户手机号。", false, "13628173752"),
			text("country", "国家", "ISO 国家码。", true, "US"),
			text("state", "州/省", "客户所在州或省。", false, "CA"),
			text("city", "城市", "客户所在城市。", false, "Los Angeles"),
			text("address", "地址", "客户地址。", false, "123 Main St, Apt 4B"),
			text("zipcode", "邮编", "客户邮编。", false, "90001"),
		}, response},
		"customer-retrieve": {"customer-retrieve", "客户", "检索客户", "通过 customerId 查询客户。", "查询客户", "GET", "/pay-api/mer/customers/{customerId}", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("customerId", "客户 ID", "创建客户返回的 customerId。", true, "cus_xxx"),
		}, response},
		"customer-update": {"customer-update", "客户", "更新客户", "更新网关客户资料。", "更新客户", "PUT", "/pay-api/mer/customers/{customerId}", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("customerId", "客户 ID", "待更新客户 ID。", true, "cus_xxx"),
			text("firstname", "名", "客户名字。", true, "Lily"),
			text("lastname", "姓", "客户姓氏。", true, "Brown"),
			text("email", "邮箱", "客户邮箱。", true, "lily_brown_1782457030419@test.com"),
			text("phone", "手机号", "客户手机号。", false, "13628173752"),
			text("country", "国家", "ISO 国家码。", true, "US"),
			text("state", "州/省", "客户所在州或省。", false, "CA"),
			text("city", "城市", "客户所在城市。", false, "San Francisco"),
			text("address", "地址", "客户地址。", false, "123 Main St, Apt 4B"),
			text("zipcode", "邮编", "客户邮编。", false, "90001"),
		}, response},
		"customer-delete": {"customer-delete", "客户", "删除客户", "删除网关客户资料。", "删除客户", "DELETE", "/pay-api/mer/customers/{customerId}", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("customerId", "客户 ID", "待删除客户 ID。", true, "cus_xxx"),
		}, response},
		"customer-list": {"customer-list", "客户", "列出客户", "列出网关客户资料。", "列出客户", "GET", "/pay-api/mer/customers", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
		}, response},
		"payin-checkout": {"payin-checkout", "代收", "创建收银台代收", "创建跳转式代收交易。", "创建支付", "POST", "/pay-api/trade/payment", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("orderNo", "商户订单号", "Demo 自动生成。", true, "AUTO_PAYIN_CHECKOUT"),
			selectField("currency", "币种", "交易币种。", true, "USD", currency),
			text("amount", "金额", "主币种金额，使用字符串。", true, "12.34"),
			text("returnUrl", "返回地址", "支付完成后的跳转地址。", false, "https://manage.forgottenthrone.com/"),
			text("notifyUrl", "异步通知地址", "网关支付结果通知地址。", false, "http://127.0.0.1:58083/webhook/payin"),
			text("clientIp", "客户端 IP", "付款人客户端 IP。", false, "47.125.221.223"),
			text("website", "商户网站", "商户站点。", false, "https://manage.forgottenthrone.com/"),
			arraySelect("paymentMethodTypes", "可用支付方式", "收银台代收提交为数组。", false, "CARD", methods),
			selectField("customerMode", "客户提交方式", "customerId 与 customer 二选一。", false, "customer", []string{"customer", "customerId"}),
			text("customerId", "客户 ID", "已有客户 ID。", false, ""),
			jsonField("customer", "客户信息", "内联客户资料。", false, "CUSTOMER"),
		}, response},
		"payin-direct": {"payin-direct", "代收", "创建直连代收", "切换支付方式会替换 paymentMethodData。", "发起支付", "POST", "/pay-api/trade/payment", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("orderNo", "商户订单号", "Demo 自动生成。", true, "AUTO_PAYIN_DIRECT"),
			selectField("payType", "支付类型", "直连代收固定为 1。", true, "1", []string{"1"}),
			selectField("currency", "币种", "交易币种。", true, "USD", currency),
			text("amount", "金额", "主币种金额，使用字符串。", true, "12.34"),
			selectField("paymentMethod", "支付方式", "不同方式需要不同参数。", true, "CASHAPP", methods),
			jsonField("paymentMethodData", "支付方式参数", "随 paymentMethod 自动替换。", true, "PAYMENT_METHOD_DATA_CASHAPP"),
			text("notifyUrl", "异步通知地址", "网关支付结果通知地址。", false, "http://127.0.0.1:58083/webhook/payin"),
			selectField("customerMode", "客户提交方式", "customerId 与 customer 二选一。", false, "customer", []string{"customer", "customerId"}),
			text("customerId", "客户 ID", "已有客户 ID。", false, ""),
			jsonField("customer", "客户信息", "内联客户资料。", false, "CUSTOMER"),
		}, response},
		"payin-retrieve": {"payin-retrieve", "代收", "查询代收交易", "按平台代收交易号查询。", "查询支付", "GET", "/pay-api/trade/payment/{tradeNo}", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("tradeNo", "平台交易号", "创建代收返回的 tradeNo。", true, "pay_xxx"),
		}, response},
		"refund-create": {"refund-create", "退款申请", "创建退款申请", "基于代收交易发起退款。", "申请退款", "POST", "/pay-api/trade/refund", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("tradeNo", "原代收交易号", "需要退款的 tradeNo。", true, "pay_xxx"),
			selectField("currency", "币种", "必须与原交易一致。", true, "USD", currency),
			text("amount", "原交易金额", "原交易金额。", true, "12.34"),
			text("refundAmount", "退款金额", "本次退款金额。", true, "1.00"),
			text("refundReason", "退款原因", "提交给网关的原因。", true, "Go SDK demo refund request"),
		}, response},
		"refund-retrieve": {"refund-retrieve", "退款申请", "查询退款", "按退款标识查询。", "查询退款", "GET", "/pay-api/trade/refund/{refundNo}", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("refundNo", "退款标识", "退款申请返回的 charge/refundNo。", true, "charge_xxx"),
		}, response},
		"payout-create": {"payout-create", "代付", "发起代付", "创建代付申请。", "发起代付", "POST", "/pay-api/payout/trade/transfer", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("orderNo", "商户订单号", "Demo 自动生成。", true, "AUTO_PAYOUT"),
			selectField("currency", "币种", "代付币种。", true, "USD", currency),
			text("amount", "金额", "出款金额。", true, "3.11"),
			selectField("paymentMethod", "支付方式", "收款支付方式。", true, "CARD", methods),
			jsonField("paymentMethodData", "支付方式参数", "随 paymentMethod 自动替换。", true, "PAYMENT_METHOD_DATA_CARD"),
			text("notifyUrl", "异步通知地址", "网关代付结果通知地址。", false, "http://127.0.0.1:58084/webhook/payout"),
			text("clientIp", "客户端 IP", "操作或客户 IP。", true, "47.125.221.223"),
			text("website", "商户网站", "商户站点。", false, "https://manage.forgottenthrone.com/"),
			text("metadata", "透传字段", "商户自定义透传数据。", false, "metadata"),
			jsonField("customer", "客户信息", "收款人资料。", false, "CUSTOMER"),
		}, response},
		"payout-retrieve": {"payout-retrieve", "代付", "查询代付", "按平台代付交易号查询。", "查询代付", "GET", "/pay-api/payout/trade/transfer/{tradeNo}", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("tradeNo", "平台代付交易号", "代付返回的 tradeNo。", true, "payout_xxx"),
		}, response},
		"payout-cancel": {"payout-cancel", "代付", "取消代付", "提交代付取消申请。", "取消代付", "POST", "/pay-api/payout/trade/transfer-cancel", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			text("tradeNo", "平台代付交易号", "代付返回的 tradeNo。", true, "payout_xxx"),
			text("orderNo", "商户订单号", "原代付订单号。", true, "PAYOUT_xxx"),
			text("remark", "备注", "取消原因。", false, "Go SDK demo payout cancel"),
		}, response},
		"balance-retrieve": {"balance-retrieve", "余额查询", "查询资金账户余额", "查询商户余额。", "查询余额", "GET", "/pay-api/fund/accounts/get", []DemoField{
			text("merchantNo", "商户号", "自动来自配置。", true, ""),
			selectField("currency", "币种", "为空可查询默认范围。", false, "USD", currency),
		}, response},
	}
}

func text(name, label, desc string, required bool, def string) DemoField {
	return DemoField{name, label, desc, required, "text", def, nil}
}

func jsonField(name, label, desc string, required bool, def string) DemoField {
	return DemoField{name, label, desc, required, "json", def, nil}
}

func selectField(name, label, desc string, required bool, def string, options []string) DemoField {
	return DemoField{name, label, desc, required, "select", def, options}
}

func arraySelect(name, label, desc string, required bool, def string, options []string) DemoField {
	return DemoField{name, label, desc, required, "array", def, options}
}
