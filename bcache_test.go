package bcache_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/gavrilaf/bcache"
)

type BCacheTestSuite struct {
	suite.Suite

	redisMock   *miniredis.Miniredis
	redisClient *redis.Client
}

func (suite *BCacheTestSuite) SetupTest() {
	t := suite.T()

	var err error
	suite.redisMock, err = miniredis.Run()
	require.NoError(t, err)

	suite.redisClient = redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    suite.redisMock.Addr(),
	})
}

func (suite *BCacheTestSuite) TearDownTest() {
	suite.redisClient.Close()
	suite.redisMock.Close()
}

func (suite *BCacheTestSuite) TestBCache() {
	t := suite.T()

	tests := []struct {
		name  string
		cache bcache.Client
	}{
		{
			name: "bigcache + buffered msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder: bcache.NewBufferedMsgPackCoder(),
				Local: bcache.NewBigCacheCache(),
				Remote: &bcache.RedisCache{
					TTL:    time.Minute * 10,
					Client: suite.redisClient,
				},
			}),
		},
	}

	obj := TestType{
		I:  123,
		S:  "test",
		A:  []int{2, 3, 4},
		M:  map[string]string{"1": "2"},
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

func TestBCacheTestSuite(t *testing.T) {
	suite.Run(t, new(BCacheTestSuite))
}

func BenchmarkBCache(b *testing.B) {
	redisMock, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	defer redisMock.Close()

	redisClient := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    redisMock.Addr(),
	})
	defer redisClient.Close()

	remote := &bcache.RedisCache{
		TTL:    time.Minute * 10,
		Client: redisClient,
	}

	benchmarks := []struct {
		name  string
		cache bcache.Client
	}{
		{
			name: "lru + msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder:  bcache.VanillaMsgPackCoder{},
				Local:  bcache.NewSimpleLRUCache(),
				Remote: remote,
			}),
		},
		{
			name: "lru + buffered msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder:  bcache.NewBufferedMsgPackCoder(),
				Local:  bcache.NewSimpleLRUCache(),
				Remote: remote,
			}),
		},
		{
			name: "bigcache + buffered msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder:  bcache.NewBufferedMsgPackCoder(),
				Local:  bcache.NewBigCacheCache(),
				Remote: remote,
			}),
		},
		{
			name: "bigcache + msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder:  bcache.VanillaMsgPackCoder{},
				Local:  bcache.NewBigCacheCache(),
				Remote: remote,
			}),
		},
		{
			name: "ristretto + json",
			cache: bcache.NewClient(bcache.Config{
				Coder:  bcache.JsonCoder{},
				Local:  bcache.NewRistrettoCache(),
				Remote: remote,
			}),
		},
		{
			name: "ristretto + buffered msg pack",
			cache: bcache.NewClient(bcache.Config{
				Coder:  bcache.NewBufferedMsgPackCoder(),
				Local:  bcache.NewRistrettoCache(),
				Remote: remote,
			}),
		},
	}

	obj := TestType{
		I:  123,
		S:  "test",
		A:  []int{2, 3, 4},
		M:  map[string]string{"1": "2"},
		SS: TestTypeInternal{II: 897},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				key1 := rand.Intn(100)
				_ = bb.cache.Set(context.TODO(), strconv.Itoa(key1), obj)


				key2 := rand.Intn(10)

				var cached TestType
				_ = bb.cache.Get(context.TODO(), strconv.Itoa(key2), &cached)
			}
		})
	}
}
