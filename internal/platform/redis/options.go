package redis

import "time"

type Option func(*options)

type options struct {
	poolSize        int
	minIdleConns    int
	connMaxIdleTime time.Duration

	dialTimeout   time.Duration
	readTimeout   time.Duration
	writeTimeout  time.Duration
	healthTimeout time.Duration
}

func defaultOptions() options {
	return options{
		poolSize:        20,
		minIdleConns:    2,
		connMaxIdleTime: 30 * time.Minute,

		dialTimeout:   5 * time.Second,
		readTimeout:   3 * time.Second,
		writeTimeout:  3 * time.Second,
		healthTimeout: 5 * time.Second,
	}
}

func WithPoolSize(n int) Option {
	return func(o *options) {
		if n >= 0 {
			o.poolSize = n
		}
	}
}

func WithMinIdleConns(n int) Option {
	return func(o *options) {
		if n >= 0 {
			o.minIdleConns = n
		}
	}
}

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.connMaxIdleTime = d
		}
	}
}

func WithDialTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.dialTimeout = d
		}
	}
}

func WithReadTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.readTimeout = d
		}
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.writeTimeout = d
		}
	}
}

func WithHealthTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.healthTimeout = d
		}
	}
}
