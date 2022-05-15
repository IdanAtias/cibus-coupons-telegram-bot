package db

import (
	"cibus-coupon-telegram-bot/internal/coupon"
	"encoding/json"
	"os"
)

const couponsDir = "/tmp/coupons"
const newCouponsFile = couponsDir + "/new.json"
const usedCouponsFile = couponsDir + "/used.json"

// localDB is used for local testing purposes
type localDB struct {
	cache []*coupon.Coupon // cache holds the collection of coupons; loaded at startup
}

// NewLocalDBClient creates a new local db client
func NewLocalDBClient() (*localDB, error) {
	// Create coupon files if needed
	if err := os.MkdirAll(couponsDir, os.ModePerm); err != nil {
		return nil, err
	}
	//for _, filePath := range []string{newCouponsFile, usedCouponsFile} {
	//	if _, err := os.Lstat(filePath); errors.Is(err, os.ErrNotExist) {
	//		f, err := os.Create(filePath)
	//		if err != nil {
	//			return nil, err
	//		}
	//		f.Close()
	//	} else if err != nil {
	//		// Unexpected error
	//		return nil, err
	//	}
	//}

	return &localDB{}, nil
}

func (d *localDB) Add(c *coupon.Coupon) error {
	data, err := json.Marshal(*c)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(newCouponsFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(data); err != nil {
		return err
	}
	return nil
}
