package interfaces

import "cibus-coupon-telegram-bot/internal/coupon"

// DB is the interface for accessing the coupons database
type DB interface {
	// Add adds the coupon to the database
	Add(c *coupon.Coupon) error

	// Use marks the coupon as used
	Use(c *coupon.Coupon) error

	// List lists all available coupons
	List() ([]*coupon.Coupon, error)
}
