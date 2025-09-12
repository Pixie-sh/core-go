package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// ClientConfiguration holds the configuration for the SQS client
type ClientConfiguration struct {
	Region          string `json:"region"`
	EndpointURL     string `json:"endpoint_url"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

// SQSClient is an abstraction over the AWS SQS client
type SQSClient struct {
	*sqs.Client
}

// NewSQSClient creates a new SQSClient with the given configuration
func NewSQSClient(_ context.Context, cfg ClientConfiguration) (*SQSClient, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	if cfg.EndpointURL != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.EndpointURL)
	}

	sqsClient := sqs.NewFromConfig(awsCfg)
	return &SQSClient{Client: sqsClient}, err
}
