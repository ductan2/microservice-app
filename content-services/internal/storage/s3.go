package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config contains the configuration needed to connect to an S3 compatible service.
type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
	PresignExpires  time.Duration
}

// S3Client implements ObjectStorage backed by AWS S3 or compatible services.
type S3Client struct {
	bucket         string
	client         *s3.Client
	presignClient  *s3.PresignClient
	presignExpires time.Duration
}

// NewS3Client initialises a new S3 client using the supplied configuration.
func NewS3Client(ctx context.Context, cfg S3Config) (*S3Client, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3: bucket is required")
	}

	loadOpts := []func(*config.LoadOptions) error{}
	if cfg.Region != "" {
		region := cfg.Region
		loadOpts = append(loadOpts, func(o *config.LoadOptions) error {
			config.WithRegion(region)(o)
			return nil
		})
	}
	if cfg.Endpoint != "" {
		endpoint := cfg.Endpoint
		region := cfg.Region
		loadOpts = append(loadOpts, func(o *config.LoadOptions) error {
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, _ string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           endpoint,
						SigningRegion: region,
					}, nil
				},
			))(o)
			return nil
		})
	}
	if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" {
		accessKey := cfg.AccessKeyID
		secretKey := cfg.SecretAccessKey
		loadOpts = append(loadOpts, func(o *config.LoadOptions) error {
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			))(o)
			return nil
		})
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			if _, err := url.Parse(cfg.Endpoint); err == nil {
				o.BaseEndpoint = aws.String(cfg.Endpoint)
			}
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	presignClient := s3.NewPresignClient(client)

	return &S3Client{
		bucket:         cfg.Bucket,
		client:         client,
		presignClient:  presignClient,
		presignExpires: cfg.PresignExpires,
	}, nil
}

// PutObject uploads data to the configured bucket at the provided key.
func (c *S3Client) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if c == nil {
		return fmt.Errorf("s3: client is nil")
	}
	if key == "" {
		return fmt.Errorf("s3: key is required")
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if size >= 0 {
		input.ContentLength = aws.Int64(size)
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	_, err := c.client.PutObject(ctx, input)
	return err
}

// DeleteObject removes a key from the bucket.
func (c *S3Client) DeleteObject(ctx context.Context, key string) error {
	if c == nil {
		return fmt.Errorf("s3: client is nil")
	}
	if key == "" {
		return fmt.Errorf("s3: key is required")
	}
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// GeneratePresignedURL generates a temporary URL for accessing the object.
func (c *S3Client) GeneratePresignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if c == nil {
		return "", fmt.Errorf("s3: client is nil")
	}
	if key == "" {
		return "", fmt.Errorf("s3: key is required")
	}
	opts := func(po *s3.PresignOptions) {
		if expiresIn <= 0 {
			expiresIn = c.presignExpires
		}
		if expiresIn > 0 {
			po.Expires = expiresIn
		}
	}
	res, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, opts)
	if err != nil {
		return "", err
	}
	return res.URL, nil
}
