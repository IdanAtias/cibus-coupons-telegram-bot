package interfaces

import "cibus-coupon-telegram-bot/internal/coupon"

// DB is the interface for accessing the coupons database
type DB interface {
	// Use marks the matching coupon as used
	Use(couponID string) error

	// List lists all available coupons
	List() ([]*coupon.Coupon, error)
}
