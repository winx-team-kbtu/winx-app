package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkago.Writer
}

func NewProducer(brokers []string) (*Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	for _, broker := range brokers {
		if strings.TrimSpace(broker) == "" {
			return nil, fmt.Errorf("kafka broker is empty")
		}
	}

	return &Producer{
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Balancer:               &kafkago.LeastBytes{},
			RequiredAcks:           kafkago.RequireAll,
			AllowAutoTopicCreation: true,
			Async:                  false,
			WriteTimeout:           10 * time.Second,
			ReadTimeout:            10 * time.Second,
		},
	}, nil
}

func (p *Producer) Publish(ctx context.Context, topic, key string, payload []byte) error {
	if strings.TrimSpace(topic) == "" {
		return fmt.Errorf("kafka topic is required")
	}

	msg := kafkago.Message{
		Topic: topic,
		Value: payload,
	}
	if key != "" {
		msg.Key = []byte(key)
	}

	var err error
	for i := 0; i < 3; i++ {
		err = p.writer.WriteMessages(ctx, msg)
		if err == nil {
			return nil
		}

		if errors.Is(err, kafkago.UnknownTopicOrPartition) && i < 2 {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		return fmt.Errorf("publish kafka message: %w", err)
	}

	return fmt.Errorf("publish kafka message: %w", err)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
