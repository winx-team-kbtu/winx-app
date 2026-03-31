package mongodb

import (
	"context"
	"fmt"
	"time"

	"winx-profile/configs"

	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewClient(ctx context.Context) (*mongodriver.Client, *mongodriver.Database, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongodriver.Connect(
		options.Client().
			ApplyURI(configs.Config.DB.Mongo.URI).
			SetConnectTimeout(10 * time.Second).
			SetServerSelectionTimeout(10 * time.Second),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongodb: %w", err)
	}

	if err := client.Ping(connectCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())

		return nil, nil, fmt.Errorf("ping mongodb: %w", err)
	}

	return client, client.Database(configs.Config.DB.Mongo.Database), nil
}
