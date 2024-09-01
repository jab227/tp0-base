package utils_test

import (
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
	"testing"
	"unsafe"
)

func TestPackedSizeOf(t *testing.T) {
	value := struct {
		a uint32
		b uint32
		c uint8
	}{}

	got := utils.PackedSizeOf(value)
	want := uint(unsafe.Sizeof(value.a) + unsafe.Sizeof(value.b) + unsafe.Sizeof(value.c))
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
