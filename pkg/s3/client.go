package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsManager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pixie-sh/errors-go"
)

type Client interface {
	Upload(ctx context.Context, filePath string, fileBlob io.Reader, contentTypeOptional ...string) (string, error)
	Download(ctx context.Context, filePath string) ([]byte, error)
	Stream(ctx context.Context, filePath string) (io.ReadCloser, error)
	Delete(ctx context.Context, filePath string) error
	Copy(ctx context.Context, source string, destination string) error
}

type ClientConfiguration struct {
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	UseSSL          bool   `json:"use_ssl"`
}

type client struct {
	s3Client *s3.Client
	config   ClientConfiguration
}

func NewClient(ctx context.Context, clientConfiguration ClientConfiguration) (Client, error) {
	cfgOpts := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion(clientConfiguration.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(clientConfiguration.AccessKeyID, clientConfiguration.SecretAccessKey, "")),
	}

	if !clientConfiguration.UseSSL {
		cfgOpts = append(cfgOpts, awsConfig.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}))
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if strings.Contains(clientConfiguration.Endpoint, ".s3-accelerate.amazonaws.com") {
			o.UseAccelerate = true
		} else {
			o.BaseEndpoint = aws.String(clientConfiguration.Endpoint)
		}
	})

	return &client{
		s3Client: s3Client,
		config:   clientConfiguration,
	}, nil
}

func (c *client) Upload(ctx context.Context, filePath string, fileBlob io.Reader, contentTypeOptional ...string) (string, error) {
	uploader := awsManager.NewUploader(c.s3Client)
	var contentType *string = nil
	if len(contentTypeOptional) > 0 {
		contentType = aws.String(contentTypeOptional[0])
	}

	uploadResult, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.config.Bucket),
		Key:         aws.String(filePath),
		Body:        fileBlob,
		ContentType: contentType,
	})

	if err != nil {
		pixiecontext.GetCtxLogger(ctx).Error("error uploading file %w", err)
		return "", err
	}
	return uploadResult.Location, nil
}

func (c *client) Download(ctx context.Context, filePath string) ([]byte, error) {
	result, err := c.Stream(ctx, filePath)
	if err != nil {
		return nil, errors.NewWithError(err, "failed to download file")
	}

	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			pixiecontext.GetCtxLogger(ctx).With("error", closeErr).Error("failed to close response body from s3; %s", closeErr.Error())
		}
	}(result)

	return io.ReadAll(result)
}

func (c *client) Stream(ctx context.Context, filePath string) (io.ReadCloser, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to download file")
	}

	return result.Body, nil
}

func (c *client) Delete(ctx context.Context, filePath string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return errors.NewWithError(err, "failed to delete file")
	}

	return nil
}

func (c *client) Copy(ctx context.Context, source string, destination string) error {
	co := &s3.CopyObjectInput{
		Bucket:     aws.String(c.config.Bucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s", c.config.Bucket, source)),
		Key:        aws.String(destination),
	}

	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(source),
	})
	if err != nil {
		return errors.Wrap(err, "source file does not exist", errors.NotFoundErrorCode)
	}

	pixiecontext.GetCtxLogger(ctx).
		With("source", source).
		With("destination", destination).
		With("copy_object", *co).
		Log("copying file")

	_, err = c.s3Client.CopyObject(ctx, co)
	if err != nil {
		return errors.NewWithError(err, "failed to copy file")
	}
	return nil
}
