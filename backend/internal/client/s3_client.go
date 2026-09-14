package client

import (
	"context"
	"fmt"
	"io"

	"github.com/FTATM/software-dashboard-tool/config"
	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Client struct {
	url         string
	region      string
	accessKey   string
	secretKey   string
	imageBucket string
	prefixError string
}

func NewS3Client(s3Config config.S3) model.S3Client {
	return &s3Client{
		url:         s3Config.Url,
		region:      s3Config.Region,
		accessKey:   s3Config.AccessKey,
		secretKey:   s3Config.SecretKey,
		imageBucket: s3Config.ImageBucket,
		prefixError: "s3Client",
	}
}

func (c *s3Client) getAwsClient(ctx context.Context) (*s3.Client, error) {
	const fname = "getAwsClient"
	cfg, err := awsS3Config.LoadDefaultConfig(ctx,
		awsS3Config.WithRegion(c.region),
		awsS3Config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.accessKey, c.secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.url)
		o.UsePathStyle = true
	}), nil
}

func (c *s3Client) ListAllImages(ctx context.Context) ([]string, error) {
	const fname = "ListAllImages"
	awsClient, err := c.getAwsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	var files []string
	paginator := s3.NewListObjectsV2Paginator(awsClient, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.imageBucket),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
		}
		for _, obj := range page.Contents {
			files = append(files, *obj.Key)
		}
	}
	return files, nil
}

func (c *s3Client) DeleteImage(ctx context.Context, filename string) error {
	const fname = "DeleteImage"
	awsClient, err := c.getAwsClient(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	_, err = awsClient.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.imageBucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	return nil
}

func (c *s3Client) UploadImage(ctx context.Context, file io.Reader, filename string, contentType string) error {
	const fname = "UploadImage"
	awsClient, err := c.getAwsClient(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	_, err = awsClient.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.imageBucket),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	return nil
}

func (c *s3Client) GetImage(ctx context.Context, filename string) (io.ReadCloser, *string, error) {
	const fname = "GetImage"
	awsClient, err := c.getAwsClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	out, err := awsClient.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.imageBucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	// We return the body (to stream to the user) and the content type (for headers)
	return out.Body, out.ContentType, nil
}
