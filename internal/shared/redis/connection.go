package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"planets-server/internal/shared/config"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func Connect() (*Client, error) {
	cfg := config.GlobalConfig.Redis
	logger := slog.With("component", "redis", "operation", "connect")

	if !cfg.Enabled {
		logger.Info("Redis disabled, using in-memory fallback")
		return nil, nil
	}

	var rdb *redis.Client

	if cfg.URL != "" {
		logger.Debug("Connecting to Redis using URL")
		opts, err := redis.ParseURL(cfg.URL)
		if err != nil {
			logger.Error("Failed to parse Redis URL", "error", err)
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		rdb = redis.NewClient(opts)
	} else {
		logger.Debug("Connecting to Redis using host/port",
			"host", cfg.Host,
			"port", cfg.Port)

		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to ping Redis", "error", err)
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connection established successfully")

	return &Client{rdb}, nil
}

func (c *Client) Close() error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Close()
}
