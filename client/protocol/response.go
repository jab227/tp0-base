package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
	"github.com/pkg/errors"
)


type ResponseKind uint8

const (
	_              ResponseKind = iota
	Acknowledge                 = 0
	WinnersReady                = 1
	BettingResults              = 2
	ResponseKindCount
)

var (
	ResponseHeaderSize = utils.PackedSizeOf(ResponseHeader{})
)

type Unmarshaler interface {
	UnmarshalPayload(p []byte) error
}

type ResponseHeader struct {
	PayloadSize uint32
	Kind        ResponseKind
}

type BetResponse struct {
	Header  ResponseHeader
	Payload []byte
}

type BetAcknowledge struct {
	Status uint8
}

func (b *BetAcknowledge) UnmarshalPayload(p []byte) error {
	if len(p) != 1 {
		return fmt.Errorf("bet acknowledge: malformed response")
	}
	b.Status = p[0]
	return nil
}

type Ready struct{}

func (*Ready) UnmarshalPayload(p []byte) error {
	if len(p) != 0 {
		return fmt.Errorf("bet ready: malformed response")
	}
	return nil
}

type Winners struct {
	DNIs []string
}

func (w *Winners) UnmarshalPayload(p []byte) error {
	payload := string(p)
	dnis := strings.Split(payload, ",")
	if dnis[0] == "" {
		return nil
	}
	w.DNIs = dnis
	return nil
}

func (b *BetResponse) GetType() Unmarshaler {
	switch b.Header.Kind {
	case Acknowledge:
		return &BetAcknowledge{}
	case WinnersReady:
		return &Ready{}
	case BettingResults:
		return &Winners{}
	default:
		panic("invalid kind: should be unreachable")
	}
}

func (b *BetResponse) DecodeResponse(r io.Reader) error {
	// Make constant
	buf := make([]byte, ResponseHeaderSize)
	if err := utils.ReadAtLeast(r, buf); err != nil {
		err = errors.Wrap(err, "couldn't read response header")
		return err
	}
	if buf[0] > uint8(ResponseKindCount) {
		return fmt.Errorf("wrong kind: %v", buf[0])
	}
	b.Header.Kind = ResponseKind(buf[0])
	b.Header.PayloadSize = binary.LittleEndian.Uint32(buf[1:5])
	buf = make([]byte, b.Header.PayloadSize)
	if err := utils.ReadAtLeast(r, buf); err != nil {
		err = errors.Wrap(err, "couldn't read response payload")
		return err
	}
	b.Payload = buf
	return nil
}
