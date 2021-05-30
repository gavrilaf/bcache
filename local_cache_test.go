package bcache_test

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gavrilaf/bcache"
)

func TestLocalCache(t *testing.T) {
	tests := []struct{
		name string
		cache bcache.LocalCache
	}{
		{"ristretto", bcache.NewRistrettoCache()},
		{"simple lru", bcache.NewSimpleLRUCache()},
		{"bigcache", bcache.NewBigCacheCache()},
		{"freecache", bcache.NewFreeCacheCache()},
		{"ccache", bcache.NewCCacheCache()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := tt.cache.Get("1")
			assert.False(t, ok)

			tt.cache.Set("1", []byte("123"))
			tt.cache.Set("2", []byte("321"))

			// wait for value to pass through buffers
			time.Sleep(time.Millisecond)

			s1, ok := tt.cache.Get("1")
			assert.True(t, ok)
			assert.Equal(t, []byte("123"), s1)

			s2, ok := tt.cache.Get("2")
			assert.True(t, ok)
			assert.Equal(t, []byte("321"), s2)
		})
	}
}

func BenchmarkLocalCache(b *testing.B) {
	benchmarks := []struct{
		name string
		cache bcache.LocalCache
	}{
		{"ristretto", bcache.NewRistrettoCache()},
		{"simple lru", bcache.NewSimpleLRUCache()},
		{"bigcache", bcache.NewBigCacheCache()},
		{"freecache", bcache.NewFreeCacheCache()},
		{"ccache", bcache.NewCCacheCache()},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				key1, key2 := rand.Intn(1000), rand.Intn(1000)

				bb.cache.Set(strconv.Itoa(key1), []byte("123"))
				bb.cache.Get(strconv.Itoa(key2))
			}
		})
	}
}

