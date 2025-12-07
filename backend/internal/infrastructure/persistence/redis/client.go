package redis

import (
	"context"
	"fmt"
	"time"

	"cinemaos-backend/config"
	"cinemaos-backend/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis client
type Client struct {
	client *redis.Client
	logger *logger.Logger
}

// New creates a new redis client
func New(cfg config.RedisConfig, log *logger.Logger) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
	})

	// Ping redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info("Redis connected successfully")

	return &Client{client: client, logger: log}, nil
}

// Close closes the redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Health checks redis health
func (c *Client) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetClient returns the underlying redis client
func (c *Client) GetClient() *redis.Client {
	return c.client
}
