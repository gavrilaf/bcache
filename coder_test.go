package bcache_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gavrilaf/bcache"
)

type TestTypeInternal struct {
	II int
}

type TestType struct {
	I int
	S string
	A []int
	M map[string]string
	SS TestTypeInternal
}

func TestMsgPackCoder(t *testing.T) {
	tests := []struct{
		name string
		coder bcache.Coder
	}{
		//{"simple msgpack", bcache.VanillaMsgPackCoder{}},
		//{"simple json", bcache.JsonCoder{}},
		{"buffered msgpack", bcache.NewBufferedMsgPackCoder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("int", func(t *testing.T) {
				i := 12345

				buf, err := tt.coder.Encode(i)
				assert.NoError(t, err)

				var di int
				err = tt.coder.Decode(buf, &di)
				assert.NoError(t, err)
				assert.Equal(t, i, di)
			})

			t.Run("string", func(t *testing.T) {
				s := "12345"

				buf, err := tt.coder.Encode(s)
				assert.NoError(t, err)

				var ds string
				err = tt.coder.Decode(buf, &ds)
				assert.NoError(t, err)
				assert.Equal(t, s, ds)
			})

			t.Run("object", func(t *testing.T) {
				o := TestType{
					I:  123,
					S:  "test",
					A:  []int{2, 3, 4},
					M:  map[string]string{"1": "2"},
					SS: TestTypeInternal{II: 897},
				}

				buf, err := tt.coder.Encode(o)
				assert.NoError(t, err)

				var do TestType
				err = tt.coder.Decode(buf, &do)
				assert.NoError(t, err)
				assert.Equal(t, o, do)
			})
		})
	}
}

func BenchmarkCoder(b *testing.B) {
	obj := TestType{
		I: 123,
		S: "test",
		A: []int{2, 3, 4},
		M: map[string]string{"1": "2"},
		SS: TestTypeInternal{II: 897},
	}

	benchmarks := []struct{
		name string
		coder bcache.Coder
	}{
		{"simple msgpack", bcache.VanillaMsgPackCoder{}},
		{"simple json", bcache.JsonCoder{}},
		{"buffered msgpack", bcache.NewBufferedMsgPackCoder()},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			wg := sync.WaitGroup{}
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					for j := 0; j < 10000; j++ {
						buf, _ := bb.coder.Encode(obj)

						var decoded TestType
						_ = bb.coder.Decode(buf, &decoded)
					}
					wg.Done()
				}()
			}
			wg.Wait()
		})
	}
}