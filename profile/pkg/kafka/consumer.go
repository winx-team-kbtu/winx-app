package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"winx-profile/pkg/graylog/logger"

	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafkago.Reader
}

func NewConsumer(brokers []string, topic, groupID string) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}
	if strings.TrimSpace(groupID) == "" {
		return nil, fmt.Errorf("kafka group id is required")
	}

	return &Consumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:     brokers,
			GroupID:     groupID,
			Topic:       topic,
			StartOffset: kafkago.LastOffset,
		}),
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, []byte) error) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			logger.Log.Errorf("fetch kafka message: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := handler(ctx, msg.Value); err != nil {
			logger.Log.Errorf("handle kafka message on topic %s: %v", msg.Topic, err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			logger.Log.Errorf("commit kafka message on topic %s: %v", msg.Topic, err)
			time.Sleep(2 * time.Second)
			continue
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
