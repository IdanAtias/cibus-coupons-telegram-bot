package coupon

import "time"

// Coupon is the object representing a Cibus coupon
type Coupon struct {
	id         string    // id is the coupon's unique identifier
	vendor     string    // vendor is the shopping vendor the coupon refers to
	value      float64   // value is the amount of money available in the coupon
	expiration time.Time // expiration is the date in which the coupon will expire

	// Helpers
	index    uint      // index is a positive number that is incremented upon every coupon addition
	used     bool      // used is a flag that indicates if the coupon is already used
	usedDate time.Time // usedDate is the date in which the coupon was used
}

// NewCoupon creates a new coupon
func NewCoupon(id, vendor string, value float64, expiration time.Time, index uint) *Coupon {
	return &Coupon{
		id:         id,
		vendor:     vendor,
		value:      value,
		expiration: expiration,
		index:      index,
	}
}

// Use marks the coupon object as used
func (c *Coupon) Use() {
	c.used = true
	c.usedDate = time.Now()
}
