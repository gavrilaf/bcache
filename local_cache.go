package bcache

import (
	"time"
	"unsafe"

	"github.com/allegro/bigcache/v3"
	"github.com/dgraph-io/ristretto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/coocood/freecache"
	"github.com/gavrilaf/ccache"
)

type LocalCache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte)
	Remove(key string)
}

// simple LRU cache

type SimpleLRUCache struct {
	cache *lru.Cache
}

func NewSimpleLRUCache() LocalCache {
	cache, _ := lru.New(1000)

	return &SimpleLRUCache{cache: cache}
}

func (c *SimpleLRUCache) Get(key string) ([]byte, bool) {
	v, ok := c.cache.Get(key)
	if ok {
		return v.([]byte), true
	}
	return nil, false
}

func (c *SimpleLRUCache) Set(key string, value []byte) {
	c.cache.Add(key, value)
}

func (c *SimpleLRUCache) Remove(key string) {
	c.cache.Remove(key)
}

// ristretto cache

type RistrettoCache struct {
	cache *ristretto.Cache
}

func NewRistrettoCache() LocalCache {
	cache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	return &RistrettoCache{cache: cache}
}

func (c *RistrettoCache) Get(key string) ([]byte, bool) {
	v, ok := c.cache.Get(key)
	if ok {
		return v.([]byte), true
	}
	return nil, false
}

func (c *RistrettoCache) Set(key string, value []byte) {
	c.cache.Set(key, value, 1)
}

func (c *RistrettoCache) Remove(key string) {
	c.cache.Del(key)
}

// bigcache cache

type BigCacheCache struct {
	cache *bigcache.BigCache
}

func NewBigCacheCache() LocalCache {
	cfg := bigcache.DefaultConfig(10 * time.Minute)
	cfg.Verbose = false
	cfg.Shards = 16

	cache, _ := bigcache.NewBigCache(cfg)
	return &BigCacheCache{cache: cache}
}

func (c *BigCacheCache) Get(key string) ([]byte, bool) {
	v, err := c.cache.Get(key)
	return v, err == nil
}

func (c *BigCacheCache) Set(key string, value []byte) {
	c.cache.Set(key, value)
}

func (c *BigCacheCache) Remove(key string) {
	c.cache.Delete(key)
}


// freecache

type FreeCacheCache struct {
	cache *freecache.Cache
}

func NewFreeCacheCache() LocalCache {
	cache := freecache.NewCache(512 * 1024)
	return &FreeCacheCache{cache: cache}
}

func (c *FreeCacheCache) Get(key string) ([]byte, bool) {
	bk := *(*[]byte)(unsafe.Pointer(&key))

	v, err := c.cache.Get(bk)
	return v, err == nil
}

func (c *FreeCacheCache) Set(key string, value []byte) {
	bk := *(*[]byte)(unsafe.Pointer(&key))

	c.cache.Set(bk, value, -1)
}

func (c *FreeCacheCache) Remove(key string) {
	c.cache.Del([]byte(key))
}

// CCache

type CCacheCache struct {
	cache *ccache.Cache
}

func NewCCacheCache() LocalCache {
	cache := ccache.NewCache(1000)
	return &CCacheCache{cache: cache}
}

func (c *CCacheCache) Get(key string) ([]byte, bool) {
	v, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}

	return v.([]byte), true
}

func (c *CCacheCache) Set(key string, value []byte) {
	c.cache.Set(key, value, 0)
}

func (c *CCacheCache) Remove(key string) {
	//c.cache.
}

