package db

import (
	"cibus-coupon-telegram-bot/internal/coupon"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

const (
	couponsDir     = "/tmp/coupons"
	usedCouponsDir = couponsDir + "/used"
)

// localDB is used for local testing purposes
type localDB struct{}

// NewLocalDBClient creates a new local db client
func NewLocalDBClient() (*localDB, error) {
	// Create coupon dirs if needed
	if err := os.MkdirAll(usedCouponsDir, os.ModePerm); err != nil {
		return nil, err
	}
	return &localDB{}, nil
}

// Add creates a new file in the coupons dir with the coupon data
func (d *localDB) Add(c *coupon.Coupon) error {
	data, err := json.Marshal(*c)
	if err != nil {
		return err
	}
	f, err := os.Create(couponsDir + "/" + c.ID)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(data); err != nil {
		return err
	}
	return nil
}

// Use moves the matching coupon file to the used coupons dir
func (d *localDB) Use(couponID string) error {
	oldPath, newPath := couponsDir+"/"+couponID, usedCouponsDir+"/"+couponID
	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

// List loads all new (i.e. not in the 'used' dir) coupon files, converts them to coupon objects and returns them
func (d *localDB) List() ([]*coupon.Coupon, error) {
	// Get coupon files in the coupons dir
	couponFiles, err := ioutil.ReadDir(couponsDir)
	if err != nil {
		return nil, err
	}

	// Build & aggregate coupons
	var coupons []*coupon.Coupon
	for _, couponFile := range couponFiles {
		if couponFile.IsDir() {
			// The 'used' dir
			continue
		}

		// Load file content and build the coupon object
		couponData, err := os.ReadFile(couponsDir + "/" + couponFile.Name())
		if err != nil {
			return nil, err
		}
		var c coupon.Coupon
		if err := json.Unmarshal(couponData, &c); err != nil {
			return nil, err
		}
		coupons = append(coupons, &c)
	}

	return coupons, nil
}
