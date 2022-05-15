package db

import (
	"cibus-coupon-telegram-bot/internal/coupon"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	db, err := NewLocalDBClient()
	require.NoError(t, err)
	c := coupon.NewCoupon(
		"coupon-id",
		"vendor",
		100,
		time.Now(),
		1,
	)
	require.NoError(t, db.Add(c))
}
