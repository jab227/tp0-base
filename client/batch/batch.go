package batch

import (
	"encoding/binary"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/op/go-logging"
	"unsafe"
)

const (
	kilobytes = (1 << 10)
)

var log = logging.MustGetLogger("log")
var sizeofBetLen = int(unsafe.Sizeof(uint32(0)))

type Chunk struct {
	buf   []byte
	count int
	full  bool
}

// only for assertion
var maxChunkSize int

func newChunk(maxCount, maxSize int) Chunk {
	buf := make([]byte, 0, maxSize)
	return Chunk{
		buf: buf,
	}

}

func (c Chunk) MarshalPayload() []byte {
	return c.buf
}

type Batcher struct {
	chunks   []Chunk
	full     []int
	current  int
	maxSize  int
	maxCount int
}

func NewBatcher(maxCount, maxSize int) *Batcher {
	maxChunkSize = maxSize
	return &Batcher{maxSize: maxSize, maxCount: maxCount}
}

func (b *Batcher) Push(m protocol.Marshaler) {
	if len(b.chunks) == 0 {
		chunk := newChunk(b.maxCount, b.maxSize)
		b.chunks = append(b.chunks, chunk)
	}
	chunk := &b.chunks[b.current]

	payload := m.MarshalPayload()
	payloadLen := uint32(len(payload))
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, payloadLen)
	if len(payload)+len(lenBuf) >= b.maxSize {
		panic("chunk is not big enough")
	}

	totalLen := len(chunk.buf) + len(lenBuf) + len(payload)
	if totalLen >= b.maxSize {
		b.chunks = append(b.chunks, newChunk(b.maxCount, b.maxSize))
		b.full = append(b.full, b.current)
		b.current++
		chunk = &b.chunks[b.current]
	}
	chunk.buf = append(chunk.buf, lenBuf...)
	chunk.buf = append(chunk.buf, payload...)
	chunk.count += 1
	chunk.full = chunk.count == b.maxCount
	if chunk.full {
		b.chunks = append(b.chunks, newChunk(b.maxCount, b.maxSize))
		b.full = append(b.full, b.current)
		b.current++
		chunk = &b.chunks[b.current]
	}
}

func removeOrdered(s []Chunk, i int) []Chunk {
	return append(s[:i], s[i+1:]...)
}

func (b *Batcher) Next() (Chunk, bool) {
	if len(b.full) == 0 {
		return Chunk{}, false
	}

	last := b.full[len(b.full)-1]
	b.full = b.full[:len(b.full)-1]
	chunk := b.chunks[last]
	b.chunks = removeOrdered(b.chunks, last)
	b.current--
	return chunk, true
}

func (b *Batcher) Flush() ([]Chunk, bool) {
	if len(b.chunks) == 0 {
		return nil, false
	}
	chunks := make([]Chunk, 0,len(b.full)+1)
	currentChunk := b.chunks[b.current]
	chunks = append(chunks, currentChunk)
	for i := range b.full {
		chunks = append(chunks, b.chunks[i])
	}
	b.chunks = nil
	b.full = nil
	return chunks, true
}
