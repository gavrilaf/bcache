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
	Coder  Coder
	Local  LocalCache
	Remote RemoteCache
}

func NewClient(cfg Config) Client {
	return &client{
		coder:  cfg.Coder,
		local:  cfg.Local,
		remote: cfg.Remote,
	}
}

// impl

type client struct {
	coder  Coder
	local  LocalCache
	remote RemoteCache
}

func (c *client) Get(ctx context.Context, key string, value interface{}) error {
	var buf []byte

	if c.local != nil {
		buf, _ = c.local.Get(key)
	}

	if buf == nil {
		var err error
		buf, err = c.remote.Get(ctx, key)
		if err != nil {
			return fmt.Errorf("failed to read value from remote cache, %w", err)
		}
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

	err = c.remote.Set(ctx, key, buf)
	if err == nil {
		if c.local != nil {
			c.local.Set(key, buf)
		}
	}

	return err
}

func (c *client) Remove(ctx context.Context, key string) error {
	c.local.Remove(key)
	return c.remote.Remove(ctx, key)
}
