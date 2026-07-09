package paymentgateway

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// GenerateOrderNo creates a timestamp-based demo order number with a random
// six-digit suffix. Merchants may replace this with their own id generator.
func GenerateOrderNo(prefix string) string {
	now := time.Now().UTC().Format("20060102150405000")
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%s%s000000", prefix, now)
	}
	return fmt.Sprintf("%s%s%06d", prefix, now, n.Int64())
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
