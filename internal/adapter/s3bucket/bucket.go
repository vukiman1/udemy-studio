package s3bucket

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
)

const signingRegion = "auto"

var (
	missingObjectCodes = map[string]bool{"NoSuchKey": true, "NotFound": true}
	missingBucketCodes = map[string]bool{"NoSuchBucket": true}
	deniedCodes        = map[string]bool{
		"AccessDenied": true, "InvalidAccessKeyId": true, "SignatureDoesNotMatch": true, "Unauthorized": true, "Forbidden": true,
	}
)

type Bucket struct {
	client *s3.Client
	name   string
}

func Open(connection objectstore.Connection) objectstore.Bucket {
	keys := connection.Credentials
	client := s3.New(s3.Options{
		BaseEndpoint:               aws.String(connection.Location.Endpoint),
		Region:                     signingRegion,
		UsePathStyle:               true,
		Credentials:                credentials.NewStaticCredentialsProvider(keys.AccessKeyID, keys.SecretAccessKey.Reveal(), ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &Bucket{client: client, name: connection.Location.Bucket}
}

func (bucket *Bucket) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := bucket.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket.name), Key: aws.String(key)})
	if err != nil {
		return nil, classify(err, "read "+key)
	}
	return output.Body, nil
}

func (bucket *Bucket) Put(ctx context.Context, key string, body []byte) error {
	_, err := bucket.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket.name),
		Key:           aws.String(key),
		Body:          bytes.NewReader(body),
		ContentLength: aws.Int64(int64(len(body))),
	})
	return classify(err, "write "+key)
}

func (bucket *Bucket) Delete(ctx context.Context, key string) error {
	_, err := bucket.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(bucket.name), Key: aws.String(key)})
	return classify(err, "delete "+key)
}

func classify(err error, action string) error {
	if err == nil {
		return nil
	}
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		return classifyAPIError(apiError, action, err)
	}
	var responseError *smithyhttp.ResponseError
	if errors.As(err, &responseError) && isDeniedStatus(responseError.HTTPStatusCode()) {
		return fmt.Errorf("%w: %s: %w", objectstore.ErrAccessDenied, action, err)
	}
	return fmt.Errorf("%w: %s: %w", objectstore.ErrUnreachable, action, err)
}

func classifyAPIError(apiError smithy.APIError, action string, err error) error {
	code := apiError.ErrorCode()
	switch {
	case missingObjectCodes[code]:
		return fmt.Errorf("%w: %s", objectstore.ErrObjectNotFound, action)
	case missingBucketCodes[code]:
		return fmt.Errorf("%w: %s", objectstore.ErrBucketNotFound, action)
	case deniedCodes[code]:
		return fmt.Errorf("%w: %s: %s", objectstore.ErrAccessDenied, action, apiError.ErrorMessage())
	default:
		return fmt.Errorf("%s: %w", action, err)
	}
}

func isDeniedStatus(status int) bool {
	return status == http.StatusUnauthorized || status == http.StatusForbidden
}
