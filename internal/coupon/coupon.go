package coupon

import "time"

// Coupon is the object representing a Cibus coupon
type Coupon struct {
	ID         string    `json:"id"`         // ID is the coupon's unique identifier
	Vendor     string    `json:"vendor"`     // Vendor is the shopping Vendor the coupon refers to
	Value      float64   `json:"value"`      // Value is the amount of money available in the coupon
	Expiration time.Time `json:"expiration"` // Expiration is the date in which the coupon will expire

	// Helpers
	Index    uint      `json:"index"`    // Index is a positive number that is incremented upon every coupon addition
	Used     bool      `json:"used"`     // Used is a flag that indicates if the coupon is already Used
	UsedDate time.Time `json:"usedDate"` // UsedDate is the date in which the coupon was Used
}

// NewCoupon creates a new coupon
func NewCoupon(id, vendor string, value float64, expiration time.Time, index uint) *Coupon {
	return &Coupon{
		ID:         id,
		Vendor:     vendor,
		Value:      value,
		Expiration: expiration,
		Index:      index,
	}
}

// Use marks the coupon object as Used
func (c *Coupon) Use() {
	c.Used = true
	c.UsedDate = time.Now()
}
