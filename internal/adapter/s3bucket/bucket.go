package s3bucket

import (
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

func (bucket *Bucket) Link(ctx context.Context, request objectstore.LinkRequest) (string, error) {
	input := &s3.GetObjectInput{Bucket: aws.String(bucket.name), Key: aws.String(request.Key)}
	if request.ContentDisposition != "" {
		input.ResponseContentDisposition = aws.String(request.ContentDisposition)
	}
	signed, err := s3.NewPresignClient(bucket.client).PresignGetObject(ctx, input, s3.WithPresignExpires(request.ExpiresIn))
	if err != nil {
		return "", fmt.Errorf("sign a link to %s: %w", request.Key, err)
	}
	return signed.URL, nil
}

func (bucket *Bucket) Put(ctx context.Context, upload objectstore.Upload) error {
	_, err := bucket.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket.name),
		Key:           aws.String(upload.Key),
		Body:          upload.Body,
		ContentLength: aws.Int64(upload.SizeBytes),
		ContentType:   aws.String(upload.ContentType),
	})
	return classify(err, "write "+upload.Key)
}

func (bucket *Bucket) List(ctx context.Context, prefix string) ([]objectstore.Object, error) {
	var objects []objectstore.Object
	pages := s3.NewListObjectsV2Paginator(bucket.client, &s3.ListObjectsV2Input{Bucket: aws.String(bucket.name), Prefix: aws.String(prefix)})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, classify(err, "list "+prefix)
		}
		for _, item := range page.Contents {
			objects = append(objects, objectstore.Object{Key: aws.ToString(item.Key), SizeBytes: aws.ToInt64(item.Size), ModifiedAt: aws.ToTime(item.LastModified)})
		}
	}
	return objects, nil
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
