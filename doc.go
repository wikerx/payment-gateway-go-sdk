// Package paymentgateway provides the merchant server-side SDK for integrating
// with Payment Gateway OpenAPI.
//
// The SDK handles three protocol concerns for merchant servers:
//   - signing every request with an HS256 merchant JWT;
//   - encrypting request bodies into the compact payload data format;
//   - decrypting encrypted response and webhook data.
//
// This package must only be used on trusted merchant backend services. Do not
// embed it in browsers, mobile apps, desktop apps, or any runtime where merchant
// API secrets, RSA private keys, card numbers, or CVC values may be exposed.
package paymentgateway
