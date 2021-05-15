package bcache

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type Coder interface {
	Decode(buf []byte, value interface{}) error
	Encode(value interface{}) ([]byte, error)
}

// simple msgpack

type VanillaMsgPackCoder struct{}

func (VanillaMsgPackCoder) Decode(buf []byte, value interface{}) error {
	return msgpack.Unmarshal(buf, value)
}

func (VanillaMsgPackCoder) Encode(value interface{}) ([]byte, error) {
	return msgpack.Marshal(value)
}

// simple json

type JsonCoder struct{}

func (JsonCoder) Decode(buf []byte, value interface{}) error {
	return json.Unmarshal(buf, value)
}

func (JsonCoder) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// buffered msgpack

type BufferedMsgPackCoder struct {
	pool *sync.Pool
}

func NewBufferedMsgPackCoder() Coder {
	return &BufferedMsgPackCoder{
		pool: &sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		}}
}

func (b *BufferedMsgPackCoder) Decode(buf []byte, value interface{}) error {
	return msgpack.Unmarshal(buf, value)
}

func (b *BufferedMsgPackCoder) Encode(value interface{}) ([]byte, error) {
	buf := b.pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		b.pool.Put(buf)
	}()

	e := msgpack.GetEncoder()
	defer msgpack.PutEncoder(e)
	e.Reset(buf)
	e.UseCompactInts(true)

	err := e.Encode(value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
