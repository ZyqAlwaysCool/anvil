package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NewMongo 创建 Mongo 客户端并执行 Ping。
func NewMongo(cfg config.MongoConfig) (*mongo.Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongo uri is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect failed: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping failed: %w", err)
	}
	return client, nil
}

func CloseMongo(client *mongo.Client) error {
	if client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Disconnect(ctx)
}

func TaskCollection(client *mongo.Client, database string) *mongo.Collection {
	return client.Database(database).Collection("tasks")
}
