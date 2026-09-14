package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     int
	SSLMode  string
	User     string
	Password string
	DB       string
}

func (c Config) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.DB,
	}

	query := u.Query()
	query.Set("sslmode", sslMode)
	u.RawQuery = query.Encode()

	return u.String()
}

type Client struct {
	Pool *pgxpool.Pool
	opts options
}

func NewClient(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	pgxConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	pgxConfig.MaxConns = o.maxConns
	pgxConfig.MinConns = o.minConns
	pgxConfig.MaxConnLifetime = o.maxConnLifetime
	pgxConfig.MaxConnIdleTime = o.maxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("init pool: %w", err)
	}

	client := &Client{
		Pool: pool,
		opts: o,
	}

	if err := client.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("initial ping failed: %w", err)
	}

	return client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, c.opts.healthTimeout)
	defer cancel()

	if err := c.Pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}

	return nil
}

func (c *Client) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}
