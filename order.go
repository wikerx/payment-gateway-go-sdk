package paymentgateway

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateOrderNo creates a fresh merchant order number for create-payment,
// create-payout, and other requests that are persisted by the gateway.
//
// The returned value is not a JWT jti and is not a gateway trade number. It is
// the merchant-side business idempotency number, so merchants should generate a
// new value before every create request and store it with their own order.
// Reusing a previous orderNo can be rejected by the gateway as duplicate.
//
// Merchants may replace this helper with their own database-backed sequence,
// snowflake id, or business order number generator. The SDK examples use this
// helper only to keep sample requests runnable without manual editing.
func GenerateOrderNo(prefix string) string {
	random := make([]byte, 8)
	now := time.Now().UTC().UnixNano()
	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf("%s%d", prefix, now)
	}
	return fmt.Sprintf("%s%d_%s", prefix, now, hex.EncodeToString(random))
}

// uniqueJWTID creates a per-request JWT jti value for replay protection. It is
// not a business order number and should not be used for reconciliation.
func uniqueJWTID(prefix string) string {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf("%s%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s%d_%s", prefix, time.Now().UTC().UnixNano(), hex.EncodeToString(random))
}
