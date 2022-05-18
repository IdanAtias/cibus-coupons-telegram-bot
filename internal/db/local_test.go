package db

import (
	"cibus-coupon-telegram-bot/internal/coupon"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	defer os.RemoveAll(couponsDir)
	db, err := NewLocalDBClient()
	require.NoError(t, err)
	c := coupon.NewCoupon(
		"coupon-id",
		"vendor",
		100,
		time.Now().Unix(),
	)
	require.NoError(t, db.Add(c))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.NoError(t, err)

	// Addition of identical coupons should succeed (override existing file)
	require.NoError(t, db.Add(c))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.NoError(t, err)
}

func TestUse(t *testing.T) {
	defer os.RemoveAll(couponsDir)
	db, err := NewLocalDBClient()
	require.NoError(t, err)
	c := coupon.NewCoupon(
		"coupon-id",
		"vendor",
		100,
		time.Now().Unix(),
	)
	// Add
	require.NoError(t, db.Add(c))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.NoError(t, err)

	// Use
	require.NoError(t, db.Use(c))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.True(t, errors.Is(err, os.ErrNotExist))
	_, err = os.Lstat(usedCouponsDir + "/" + "coupon-id")
	require.NoError(t, err)

	// Use again and fail
	require.Error(t, db.Use(c))

	// Use non-existing coupon and fail
	c = coupon.NewCoupon(
		"coupon-id-1",
		"vendor",
		100,
		time.Now().Unix(),
	)
	require.Error(t, db.Use(c))
}

func TestList(t *testing.T) {
	defer os.RemoveAll(couponsDir)
	db, err := NewLocalDBClient()
	require.NoError(t, err)
	coupons := []*coupon.Coupon{
		{
			ID:         "cid1",
			Vendor:     "vendor1",
			Value:      100,
			Expiration: 111,
		},
		{
			ID:         "cid2",
			Vendor:     "vendor2",
			Value:      50,
			Expiration: 222,
		},
		{
			ID:         "cid3",
			Vendor:     "vendor3",
			Value:      40,
			Expiration: 333,
		},
	}

	// Add the coupons
	for _, c := range coupons {
		require.NoError(t, db.Add(c))
	}

	// Verify 'List' returns the same coupons
	couponsList, err := db.List()
	require.NoError(t, err)
	require.ElementsMatch(t, coupons, couponsList)
}
