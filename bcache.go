package bcache

import (
	"context"
	"errors"
	"fmt"
)

var ErrCacheMiss = errors.New("key miss")


type Client interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}) error
	Remove(ctx context.Context, key string) error
}

type Config struct {
	Coder Coder
	Local LocalCache
}

func NewClient(cfg Config) Client {
	return &client{
		coder: cfg.Coder,
		local: cfg.Local,
	}
}

// impl

type client struct {
	coder Coder
	local LocalCache
}

func (c *client) Get(ctx context.Context, key string, value interface{}) error {
	var buf []byte
	if c.local != nil {
		buf, _ = c.local.Get(key)
	}

	if buf == nil {
		return ErrCacheMiss
	}

	return c.coder.Decode(buf, value)
}

func (c *client) Set(ctx context.Context, key string, value interface{}) error {
	buf, err := c.coder.Encode(value)
	if err != nil {
		return fmt.Errorf("failed to encode value, %w", err)
	}

	if c.local != nil {
		c.local.Set(key, buf)
	}

	return nil
}

func (c *client) Remove(ctx context.Context, key string) error {
	return nil
}

