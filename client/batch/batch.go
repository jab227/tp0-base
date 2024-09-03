package batch

import (
	"encoding/binary"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"unsafe"
)

const (
	kilobytes = (1 << 10)
)

var sizeofBetLen = int(unsafe.Sizeof(uint32(0)))

type Chunk struct {
	buf      []byte
	count    int
	maxCount int
	full     bool
}

// only for assertion
var maxChunkSize int

func assertMaxSize(c *Chunk) {
	if c.capacity() != maxChunkSize {
		panic("capacity of the chunk must be always maxChunkSize")
	}
}

func newChunk(maxCount, maxSize int) Chunk {
	buf := make([]byte, maxCount, maxSize)
	return Chunk{
		buf: buf, maxCount: maxCount,
	}
}

func (c *Chunk) size() int {
	return len(c.buf)
}

func (c *Chunk) capacity() int {
	return cap(c.buf)
}

func (c *Chunk) tryPush(b []byte) bool {
	currSize := c.size()
	currCapacity := c.capacity()
	if c.count == c.maxCount || currSize+len(b)+sizeofBetLen >= currCapacity {
		return false
	}
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(a, uint32(len(b)))
	c.buf = append(c.buf, a...)
	c.buf = append(c.buf, b...)
	c.count += 1
	assertMaxSize(c)
	return true
}

func (c Chunk) MarshalPayload() []byte {
	return c.buf
}

type Batcher struct {
	chunks   []Chunk
	length   int
	next     int
	maxSize  int
	maxCount int
}

func NewBatcher(maxCount, maxSize int) *Batcher {
	maxChunkSize = maxSize
	return &Batcher{maxSize: maxSize, maxCount: maxCount}
}

func (b *Batcher) Push(bet agency.Bet) {
	if b.length == 0 {
		b.addChunk()
	}
	betBytes := bet.MarshalPayload()
	lastChunk := b.getLast()
	if ok := lastChunk.tryPush(betBytes); !ok {
		b.next = b.length - 1
		lastChunk.full = true
		b.addChunk()
		lastChunk = b.getLast()
		if ok := lastChunk.tryPush(betBytes); !ok {
			panic("bigger batch size required")
		}
	}
}

func (b *Batcher) addChunk() {
	c := newChunk(b.maxCount, b.maxSize)
	b.chunks = append(b.chunks, c)
	b.length += 1
}

func removeUnordered(s []Chunk, i int) []Chunk {
	return append(s[:i], s[i+1:]...)
}

func (b *Batcher) Next() (Chunk, bool) {
	if b.length == 0 {
		return Chunk{}, false
	}
	nextIdx := b.next
	nextChunk := b.chunks[nextIdx]
	b.length = b.length - 1
	b.chunks = removeUnordered(b.chunks, nextIdx)
	idx := -1
	for i, c := range b.chunks {
		if c.full {
			idx = i
			break
		}
	}
	if idx < 0 {
		return Chunk{}, false
	}
	return nextChunk, true
}

func (b *Batcher) getLast() *Chunk {
	if b.length == 0 {
		return nil
	}
	nextIdx := b.length - 1
	return &b.chunks[nextIdx]
}
