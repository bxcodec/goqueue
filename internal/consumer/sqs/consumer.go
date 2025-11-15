package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/internal/consumer"
	consumerOpts "github.com/bxcodec/goqueue/options/consumer"
)

type sqsHandler struct {
	option *consumerOpts.ConsumerOption
}

func NewConsumer(opts ...consumerOpts.ConsumerOptionFunc) consumer.Consumer {
	opt := consumerOpts.DefaultConsumerOption()
	for _, o := range opts {
		o(opt)
	}

	sqsAWSConf, err := config.LoadDefaultConfig(context.Background(),
		config.WithDefaultRegion(cfg.SQSRegion),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           *cfg.SQSEndpoint,
				SigningRegion: cfg.SQSRegion,
			}, nil
		}),
		),
	)
	if err != nil {
		return nil, err
	}

	return &sqsHandler{
		option: opt,
	}
}

func (s *sqsHandler) Consume(ctx context.Context, handler interfaces.InboundMessageHandler, meta map[string]interface{}) (err error) {
	return nil
}

func (s *sqsHandler) Stop(ctx context.Context) (err error) {
	return nil
}
