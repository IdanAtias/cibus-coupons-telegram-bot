package coupon

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
