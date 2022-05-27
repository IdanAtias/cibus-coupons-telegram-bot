package db

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/idanatias/cibus-coupons-telegram-bot/pkg/coupon"
)

type s3Client struct {
	*s3.S3
	couponsBucket string
}

// NewS3Client returns a new s3 client
func NewS3Client(couponsBucket string) (*s3Client, error) {
	// Selected region is derived from the AWS_REGION env var
	// Note that AWS_REGION should match the region of the coupons bucket
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &s3Client{
		S3:            s3.New(sess),
		couponsBucket: couponsBucket,
	}, nil
}

// Use marks the matching coupon as used by moving it to the 'used' folder
// If coupon is not new ErrCouponAlreadyUsed is returned
// If coupon is not found ErrCouponNotExist is returned
//
// Note that the s3 client doesn't offer a Move method
// Instead, it is needed first to copy the object to the new location and then delete it
// Since this is not atomic, we could end up with a used coupon that still resides in the 'new' folder (in case 'delete' phase fails)
//
// Therefore, both 'new' and 'used' folders are checked for coupon existence
// If coupon is found in 'used', consider it used and delete it from the 'new' folder if it exists there too
// If coupon is found in 'new', consider it new and move (copy & delete) it to the 'used' folder
// Else, coupon doesn't exist
//
func (s *s3Client) Use(couponID string) error {
	// Construct a helper func that checks object existence
	isExists := func(key string) (bool, error) {
		_, err := s.GetObject(&s3.GetObjectInput{
			Bucket: &s.couponsBucket,
			Key:    &key,
		})
		if err != nil {
			// Cast err to awserr.Error to get the Code and the Message
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchKey:
					return false, nil
				default:
					log.Print(aerr.Error())
					return false, aerr
				}
			}
			log.Print(err.Error())
			return false, err
		}
		return true, nil
	}

	// Check 'used' and 'new' folders
	usedKey, newKey := fmt.Sprintf("used/%s", couponID), fmt.Sprintf("new/%s", couponID)
	isUsed, err := isExists(usedKey)
	if err != nil {
		return err
	}
	isNew, err := isExists(newKey)
	if err != nil {
		return err
	}

	if isUsed {
		log.Printf("Copuon %q is already used", couponID)
		if isNew {
			// Coupon is used but resides also in the 'new' folder
			// Delete it from the 'new' folder
			log.Printf("Used coupon %q found in the 'new' folder. Deleting it", couponID)
			if _, err := s.DeleteObject(&s3.DeleteObjectInput{
				Bucket: &s.couponsBucket,
				Key:    &newKey,
			}); err != nil {
				return err
			}
		}
		return ErrCouponAlreadyUsed
	}

	if isNew {
		// New coupon - move (copy & delete) it to the 'used' folder
		log.Printf("Moving coupon %q to the 'used' folder", couponID)
		if _, err := s.CopyObject(&s3.CopyObjectInput{
			Bucket:     &s.couponsBucket,
			Key:        &usedKey,
			CopySource: aws.String(fmt.Sprintf("%s/%s", s.couponsBucket, newKey)),
		}); err != nil {
			return err
		}
		if _, err := s.DeleteObject(&s3.DeleteObjectInput{
			Bucket: &s.couponsBucket,
			Key:    &newKey,
		}); err != nil {
			return err
		}
		return nil
	}

	// Coupon is not used nor new - it doesn't exist
	log.Printf("Coupon %q doesn't exist", couponID)
	return ErrCouponNotExist
}

// List fetches all available coupons from the coupons bucket and returns them
func (s *s3Client) List() ([]*coupon.Coupon, error) {
	// Prepare List API input
	input := &s3.ListObjectsInput{
		Bucket: &s.couponsBucket,
		Prefix: aws.String("new/"),
	}

	// Call the List API
	result, err := s.ListObjects(input)
	if err != nil {
		// Log the error and return it
		// Cast err to awserr.Error to get the Code and the Message
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf(aerr.Error())
			return nil, aerr
		}
		log.Printf(err.Error())
		return nil, err
	}

	// Go over all returned objects and create the corresponding coupon objects
	// List returns just partial object info (e.g., Key)
	// Another GET is required for getting the full object
	var coupons []*coupon.Coupon
	for _, objRef := range result.Contents {
		obj, err := s.GetObject(&s3.GetObjectInput{
			Bucket: &s.couponsBucket,
			Key:    objRef.Key,
		})
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(obj.Body)
		if err != nil {
			return nil, err
		}
		var c coupon.Coupon
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, err
		}
		coupons = append(coupons, &c)
	}

	return coupons, nil
}
