package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	s3svc "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"golang.org/x/oauth2/google"
	"io"
)

type S3 struct {
	client s3iface.S3API
}

const (
	bucketOCDTrackerAPI     = "ocdtracker-api"
	objectKeyGoogleAppCreds = "google-app-creds.json"
)

func NewS3(sess *session.Session) *S3 {
	config := &aws.Config{Region: aws.String(defaultRegion)}
	return &S3{client: s3svc.New(sess, config)}
}

func (s3 *S3) getObject(bucket, key string) ([]byte, error) {
	input := &s3svc.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}
	objectOutput, err := s3.client.GetObject(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from bucket %s: %w", key, bucket, err)
	}
	defer objectOutput.Body.Close()
	body, err := io.ReadAll(objectOutput.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from s3: %w", err)
	}
	return body, nil
}

func (s3 *S3) GetGoogleAppCreds(ctx context.Context) (*google.Credentials, error) {
	objectOutput, err := s3.getObject(bucketOCDTrackerAPI, objectKeyGoogleAppCreds)
	if err != nil {
		return nil, err
	}
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/firebase"}
	googleAppCreds, err := google.CredentialsFromJSON(ctx, objectOutput, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to build google application credentials from json: %w", err)
	}
	return googleAppCreds, nil
}
