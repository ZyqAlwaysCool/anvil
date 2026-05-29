//go:build ignore

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"anvil-scaffold-template/internal/platform/config"
)

// NewRedis 创建 Redis 客户端并执行 Ping；失败时返回带资源上下文的错误。
func NewRedis(cfg config.RedisConfig) (redis.UniversalClient, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	var client redis.UniversalClient
	switch cfg.Mode {
	case "single":
		client = redis.NewClient(&redis.Options{
			Addr:         cfg.Addrs[0],
			Username:     cfg.Username,
			Password:     cfg.Password,
			DB:           cfg.DB,
			DialTimeout:  cfg.Timeout,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
		})
	case "cluster":
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Addrs,
			Username:     cfg.Username,
			Password:     cfg.Password,
			DialTimeout:  cfg.Timeout,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
		})
	default:
		return nil, fmt.Errorf("redis: invalid mode %s", cfg.Mode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return client, nil
}

func CloseRedis(client redis.UniversalClient) error {
	if client == nil {
		return nil
	}
	return client.Close()
}

func PingRedis(ctx context.Context, client redis.UniversalClient, timeout time.Duration) error {
	if client == nil {
		return fmt.Errorf("redis client is nil")
	}
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.Ping(pingCtx).Err()
}
