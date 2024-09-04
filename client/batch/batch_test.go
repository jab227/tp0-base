package batch_test

import (
	"bytes"
	"encoding/binary"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/batch"
	"reflect"
	"testing"
)

type StubMarshaler struct {
	cnt int
}

func (s *StubMarshaler) MarshalPayload() []byte {
	s.cnt += 1
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(a, uint32(s.cnt))
	return a
}

func TestBatcher(t *testing.T) {
	const (
		MaxCount = 3
		MaxSize  = 8 * 1024
	)
	batcher := batch.NewBatcher(MaxCount, MaxSize)
	stub := StubMarshaler{}
	for i := 0; i < 3; i++ {
		batcher.Push(&stub)
	}

	var buf bytes.Buffer
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(a, uint32(4))
	buf.Write(a)
	buf.Write([]byte{0x01, 0x00, 0x00, 0x00})
	buf.Write(a)
	buf.Write([]byte{0x02, 0x00, 0x00, 0x00})
	buf.Write(a)
	buf.Write([]byte{0x03, 0x00, 0x00, 0x00})
	want := buf.Bytes()
	chunk, ok := batcher.Next()
	if !ok {
		t.Error("expected ok")
	}

	got := chunk.MarshalPayload()
	if !reflect.DeepEqual(got, []byte(want)) {
		t.Errorf("got %v, want %v", got, want)
	}
	batcher.Push(&stub)
	_, ok = batcher.Next()
	if ok {
		t.Error("expected not ok")
	}
}
