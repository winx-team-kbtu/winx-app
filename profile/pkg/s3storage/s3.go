package s3storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"winx-profile/configs"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	SignedGetURL(ctx context.Context, key string, expires time.Duration) (string, error)
	Bucket() string
}

type Client struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	region        string
}

func NewClient(ctx context.Context) (*Client, error) {
	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(configs.Config.S3.Region),
	}

	if configs.Config.S3.AccessKeyID != "" && configs.Config.S3.SecretAccessKey != "" {
		loadOptions = append(loadOptions,
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					configs.Config.S3.AccessKeyID,
					configs.Config.S3.SecretAccessKey,
					"",
				),
			),
		)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &Client{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        configs.Config.S3.Bucket,
		region:        configs.Config.S3.Region,
	}, nil
}

func (c *Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &c.bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("put s3 object: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucket, c.region, key), nil
}

func (c *Client) SignedGetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	req, err := c.presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: &c.bucket,
			Key:    &key,
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = expires
		},
	)
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}

	return req.URL, nil
}

func (c *Client) Bucket() string {
	return c.bucket
}
