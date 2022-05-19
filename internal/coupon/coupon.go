package coupon

import (
	"fmt"
	"time"
)

// Coupon is the object representing a Cibus coupon
type Coupon struct {
	ID         string `json:"id"`         // ID is the coupon's unique identifier
	Vendor     string `json:"vendor"`     // Vendor is the shopping Vendor the coupon refers to
	Value      int    `json:"value"`      // Value is the amount of money available in the coupon
	Expiration int64  `json:"expiration"` // Expiration is the epoch timestamp in which the coupon will expire
}

// NewCoupon creates a new coupon
func NewCoupon(id, vendor string, value int, expiration int64) *Coupon {
	return &Coupon{
		ID:         id,
		Vendor:     vendor,
		Value:      value,
		Expiration: expiration,
	}
}

// String returns a string representation of the coupon
func (c *Coupon) String() string {
	return fmt.Sprintf(
		"%s | %vILS | %s | %s",
		c.ID,
		c.Value,
		c.Vendor,
		time.Unix(c.Expiration, 0).Format(time.RFC822),
	)
}

// ReadableCouponID puts a dash after every 4 chars of coupon ID for making it more readable
func ReadableCouponID(couponID string) string {
	var readableCouponID string
	for i := 0; i < len(couponID); i++ {
		if i != 0 && i%4 == 0 {
			readableCouponID = readableCouponID + "-"
		}
		readableCouponID = readableCouponID + string(couponID[i])
	}
	return readableCouponID
}
