package batch_test

import (
	"fmt"
	"reflect"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/batch"
	"testing"
)

type StubMarshaler struct {
	cnt int
}

func (s *StubMarshaler) MarshalPayload() []byte {
	s.cnt += 1
	return []byte(fmt.Sprintf("payload: %d", s.cnt))
}
func TestBatcher(t *testing.T) {
	const (
		MaxCount = 3
		MaxSize  = 8 * 1024
	)
	batcher := batch.NewBatcher(MaxCount, MaxSize)
	stub := StubMarshaler{}
	for i := 0; i < 9; i++ {
		batcher.Push(&stub)
	}

	for i := 0; i < 3; i++ {
		chunk, ok := batcher.Next()
		if !ok {
			t.Error("expected ok")
		}

		want := "payload: 1payload: 2payload: 3"
		got := chunk.MarshalPayload()
		if !reflect.DeepEqual(got, []byte(want)) {
			t.Errorf("got %v, want %v", string(got), want)
		}
	}
}
