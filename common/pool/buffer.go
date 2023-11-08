package pool

import (
	"bytes"
	"sync"

	"github.com/MerlinKodo/protobytes"
)

var (
	bufferPool      = sync.Pool{New: func() any { return &bytes.Buffer{} }}
	bytesBufferPool = sync.Pool{New: func() any { return &protobytes.BytesWriter{} }}
)

func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

func GetBytesBuffer() *protobytes.BytesWriter {
	return bytesBufferPool.Get().(*protobytes.BytesWriter)
}

func PutBytesBuffer(buf *protobytes.BytesWriter) {
	buf.Reset()
	bytesBufferPool.Put(buf)
}
