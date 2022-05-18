package db

import (
	"cibus-coupon-telegram-bot/internal/coupon"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

type s3Client struct {
	*s3.S3
	couponsBucket string
}

// NewS3Client returns a new s3 client
func NewS3Client(couponsBucket string) (*s3Client, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &s3Client{
		S3:            s3.New(sess),
		couponsBucket: couponsBucket,
	}, nil
}

// Add is not implemented for s3 client as currently this is not needed
func (s *s3Client) Add(c *coupon.Coupon) error {
	return nil
}

// List fetces all available coupons from the coupons bucket and returns them
func (s *s3Client) List() ([]*coupon.Coupon, error) {
	// Prepare List API input
	input := &s3.ListObjectsInput{
		Bucket: aws.String(s.couponsBucket),
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
	fmt.Println(result.)
}
