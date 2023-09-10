package db

import (
	"errors"
	"os"
	"testing"

	"github.com/idanatias/cibus-coupons-telegram-bot/pkg/coupon"

	"github.com/stretchr/testify/require"
)

// A workaround for Go1.18
// 'testing' package introduced a breaking change:
// testing.T method signatures were modified - 'interface{}' was replaced with 'any'
// Therefore 'require' doesn't recognize testing.T as implementing 'Errorf(format string, arg ...interface{})'
//
// As there is currently no newer version for the 'testify' module with a suitable 'require' package, we'll work around
// this by creating a new type wrapper for testing.T which implements the expected interface
type myTesting struct {
	*testing.T
}

func (t *myTesting) Errorf(format string, args ...interface{}) {
	t.Errorf(format, args...)
}

func (t *myTesting) FailNow() {
	t.FailNow()
}

func TestAdd(t *testing.T) {
	mt := &myTesting{t}
	defer os.RemoveAll(couponsDir)
	db, err := NewLocalDBClient()
	require.NoError(mt, err)
	c := coupon.NewCoupon(
		"coupon-id",
		"vendor",
		100,
	)
	require.NoError(mt, db.Add(c))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.NoError(mt, err)

	// Addition of identical coupons should succeed (override existing file)
	require.NoError(mt, db.Add(c))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.NoError(mt, err)
}

func TestUse(t *testing.T) {
	mt := &myTesting{t}
	defer os.RemoveAll(couponsDir)
	db, err := NewLocalDBClient()
	require.NoError(mt, err)
	c := coupon.NewCoupon(
		"coupon-id",
		"vendor",
		100,
	)
	// Add
	require.NoError(mt, db.Add(c))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.NoError(mt, err)

	// Use
	require.NoError(mt, db.Use(c.ID))
	_, err = os.Lstat(couponsDir + "/" + "coupon-id")
	require.True(mt, errors.Is(err, os.ErrNotExist))
	_, err = os.Lstat(usedCouponsDir + "/" + "coupon-id")
	require.NoError(mt, err)

	// Use again and fail
	require.Error(mt, db.Use(c.ID))

	// Use non-existing coupon and fail
	c = coupon.NewCoupon(
		"coupon-id-1",
		"vendor",
		100,
	)
	require.Error(mt, db.Use(c.ID))
}

func TestList(t *testing.T) {
	mt := &myTesting{t}
	defer os.RemoveAll(couponsDir)
	db, err := NewLocalDBClient()
	require.NoError(mt, err)
	coupons := []*coupon.Coupon{
		{
			ID:     "cid1",
			Vendor: "vendor1",
			Value:  100,
		},
		{
			ID:     "cid2",
			Vendor: "vendor2",
			Value:  50,
		},
		{
			ID:     "cid3",
			Vendor: "vendor3",
			Value:  40,
		},
	}

	// Add the coupons
	for _, c := range coupons {
		require.NoError(mt, db.Add(c))
	}

	// Verify 'List' returns the same coupons
	couponsList, err := db.List()
	require.NoError(mt, err)
	require.ElementsMatch(mt, coupons, couponsList)
}
