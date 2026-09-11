package postgres

import "time"

type Option func(*options)

type options struct {
	maxConns        int32
	minConns        int32
	maxConnLifetime time.Duration
	maxConnIdleTime time.Duration
	healthTimeout   time.Duration
}

func defaultOptions() options {
	return options{
		maxConns:        25,
		minConns:        5,
		maxConnLifetime: 1 * time.Hour,
		maxConnIdleTime: 30 * time.Minute,
		healthTimeout:   5 * time.Second,
	}
}

func WithMaxConns(n int32) Option {
	return func(o *options) {
		if n > 0 {
			o.maxConns = n
		}
	}
}

func WithMinConns(n int32) Option {
	return func(o *options) {
		if n >= 0 {
			o.minConns = n
		}
	}
}

func WithMaxConnLifetime(d time.Duration) Option {
	return func(o *options) {
		o.maxConnLifetime = d
	}
}

func WithMaxConnIdleTime(d time.Duration) Option {
	return func(o *options) {
		o.maxConnIdleTime = d
	}
}

func WithHealthTimeout(d time.Duration) Option {
	return func(o *options) {
		o.healthTimeout = d
	}
}
