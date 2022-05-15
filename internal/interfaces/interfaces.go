package interfaces

import "cibus-coupon-telegram-bot/internal/coupon"

// DB is the interface for accessing the coupons database
type DB interface {
	// Add adds the coupon to the database
	Add(c *coupon.Coupon) error

	// GetByID gets a coupon by its id
	GetByID(id string) (error, *coupon.Coupon)

	// GetByIndex gets a coupon by its index
	GetByIndex(index uint) (error, *coupon.Coupon)

	// List lists all new coupons
	List() (error, []*coupon.Coupon)

	// ListUsed lists all used coupons
	ListUsed() (error, []*coupon.Coupon)
}
