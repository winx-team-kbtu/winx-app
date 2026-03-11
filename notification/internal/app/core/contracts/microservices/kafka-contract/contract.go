package kafkacontract

import "context"

type Producer interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
	Close() error
}
