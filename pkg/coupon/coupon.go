package coupon

import (
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"image/png"
	"os"
)

// Coupon is the object representing a Cibus coupon
type Coupon struct {
	ID     string `json:"id"`     // ID is the coupon's unique identifier
	Vendor string `json:"vendor"` // Vendor is the shopping Vendor the coupon refers to
	Value  int    `json:"value"`  // Value is the amount of money available in the coupon
}

// NewCoupon creates a new coupon
func NewCoupon(id, vendor string, value int) *Coupon {
	return &Coupon{
		ID:     id,
		Vendor: vendor,
		Value:  value,
	}
}

// String returns a string representation of the coupon
func (c *Coupon) String() string {
	return fmt.Sprintf(
		"%s | %vILS | %s",
		c.ID,
		c.Value,
		c.Vendor,
	)
}

// ReadableCouponID puts a dash after every 4 chars of coupon ID for making it more readable
func ReadableCouponID(couponID string) string {
	var readableCouponID string
	for i := 0; i < len(couponID); i++ {
		if i != 0 && i%4 == 0 {
			readableCouponID = readableCouponID + "-"
		}
		readableCouponID = readableCouponID + string(couponID[i])
	}
	return readableCouponID
}

// GenerateBarcodeFile generates a barcode png photo based on the coupon ID and return its path
func GenerateBarcodeFile(couponID string) (string, error) {
	// Create the barcode
	code, err := code128.Encode(couponID)
	if err != nil {
		return "", err
	}

	// Scale the barcode so it can be easily scanned
	scaledCode, err := barcode.Scale(code, 500, 200)
	if err != nil {
		return "", err
	}

	// Create the output file encoded as png
	filePath := "/tmp/barcode.png"
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return filePath, png.Encode(file, scaledCode)
}
