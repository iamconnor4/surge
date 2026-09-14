package redis

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (c Config) Addr() string {
	return net.JoinHostPort(
		c.Host,
		strconv.Itoa(c.Port),
	)
}

type Client struct {
	Client *redis.Client
	opts   options
}

func NewClient(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	redisOptions := &redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,

		PoolSize:        o.poolSize,
		MinIdleConns:    o.minIdleConns,
		ConnMaxIdleTime: o.connMaxIdleTime,

		DialTimeout:  o.dialTimeout,
		ReadTimeout:  o.readTimeout,
		WriteTimeout: o.writeTimeout,
	}

	rdb := redis.NewClient(redisOptions)

	client := &Client{
		Client: rdb,
		opts:   o,
	}

	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("initial ping failed: %w", err)
	}

	return client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, c.opts.healthTimeout)
	defer cancel()

	if err := c.Client.Ping(pingCtx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	if c.Client == nil {
		return nil
	}

	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("redis close failed: %w", err)
	}

	return nil
}
