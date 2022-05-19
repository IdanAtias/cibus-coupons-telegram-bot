package db

import "errors"

var (
	ErrCouponNotExist    = errors.New("coupon does not exist")
	ErrCouponAlreadyUsed = errors.New("coupon already used")
)
