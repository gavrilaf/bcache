package bcache_test

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gavrilaf/bcache"
)

func TestBCache(t *testing.T) {
	tests := []struct{
		name string
		cache bcache.Client
	}{
		{
			name: "bigcache + buffered msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder: bcache.NewBufferedMsgPackCoder(),
				Local: bcache.NewBigCacheCache(),
			}),
		},
	}

	obj := TestType{
		I: 123,
		S: "test",
		A: []int{2, 3, 4},
		M: map[string]string{"1": "2"},
		SS: TestTypeInternal{II: 897},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			c := tt.cache

			var cached TestType

			err := c.Get(ctx, "1", &cached)
			assert.Error(t, err)
			assert.EqualError(t, err, bcache.ErrCacheMiss.Error())

			err = c.Set(ctx, "1", obj)
			assert.NoError(t, err)

			err = c.Get(ctx, "1", &cached)
			assert.NoError(t, err)
			assert.Equal(t, obj, cached)
		})
	}
}

func BenchmarkBCache(b *testing.B) {
	benchmarks := []struct{
		name string
		cache bcache.Client
	}{
		{
			name: "bigcache + buffered msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder: bcache.NewBufferedMsgPackCoder(),
				Local: bcache.NewBigCacheCache(),
			}),
		},
		{
			name: "bigcache + msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder: bcache.VanillaMsgPackCoder{},
				Local: bcache.NewBigCacheCache(),
			}),
		},
		{
			name: "ristretto + json",
			cache: bcache.NewClient(bcache.Config{
				Coder: bcache.JsonCoder{},
				Local: bcache.NewRistrettoCache(),
			}),
		},
		{
			name: "ristretto + buffered msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder: bcache.NewBufferedMsgPackCoder(),
				Local: bcache.NewRistrettoCache(),
			}),
		},
		{
			name: "lru + msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder: bcache.VanillaMsgPackCoder{},
				Local: bcache.NewSimpleLRUCache(),
			}),
		},
	}

	obj := TestType{
		I: 123,
		S: "test",
		A: []int{2, 3, 4},
		M: map[string]string{"1": "2"},
		SS: TestTypeInternal{II: 897},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			wg := sync.WaitGroup{}
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					for i := 0; i < 10000; i++ {
						key1, key2 := rand.Intn(1000), rand.Intn(1000)

						_ = bb.cache.Set(context.TODO(), strconv.Itoa(key1), obj)

						var cached TestType
						_ = bb.cache.Get(context.TODO(), strconv.Itoa(key2), &cached)
					}
					wg.Done()
				}()
			}
			wg.Wait()
		})
	}
}
