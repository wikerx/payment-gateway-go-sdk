package paymentgateway

import (
	"strings"
	"testing"
	"time"
)

func TestMerchantJWTSigner(t *testing.T) {
	token, err := NewMerchantJWTSigner().Sign("2606177036", strings.Repeat("a", 32), false, "JTI_1", time.Unix(1000, 0), 180)
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(token, ".")) != 3 {
		t.Fatalf("jwt must have three parts: %s", token)
	}
}

func TestUniqueJWTIDDoesNotRepeat(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		jti := uniqueJWTID("PAYOUT_CREATE_")
		if seen[jti] {
			t.Fatalf("duplicate jwt id generated: %s", jti)
		}
		seen[jti] = true
	}
}
