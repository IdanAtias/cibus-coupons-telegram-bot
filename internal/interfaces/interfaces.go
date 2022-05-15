package interfaces

import "cibus-coupon-telegram-bot/internal/coupon"

// DB is the interface for accessing the coupons database
type DB interface {
	// Add adds the coupon to the database
	Add(c *coupon.Coupon) error

	// GetByID gets a coupon by its id
	//GetByID(id string) (*coupon.Coupon, error)
	//
	//// GetByIndex gets a coupon by its index
	//GetByIndex(index uint) (*coupon.Coupon, error)
	//
	//// List lists all new coupons
	//List() ([]*coupon.Coupon, error)
	//
	//// ListUsed lists all used coupons
	//ListUsed() ([]*coupon.Coupon, error)
}
