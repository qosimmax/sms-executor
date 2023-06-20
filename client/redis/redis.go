package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"

	"github.com/qosimmax/sms-executor/config"
)

type Client struct {
	redis *redis.Client
	topic string
}

// Init initializes a new client.
func (c *Client) Init(ctx context.Context, config *config.Config) error {
	c.redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", config.RedisAddress),
		Password: "", // no password set
		PoolSize: 20,
	})

	err := c.redis.Ping(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("error pinging redis: %w", err)
	}

	c.topic = config.NatsTopic

	return nil
}
