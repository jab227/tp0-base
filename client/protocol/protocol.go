package protocol

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// Byte sizes
const (
	ResponseSize      = 4
	RequestHeaderSize = 13
)

type Marshaler interface {
	MarshalBet() ([]byte, int)
}

type MessageKind uint8

const (
	_                   = iota
	PostBet MessageKind = 0
)

type RequestHeader struct {
	PayloadSize uint32
	Count       uint32
	AgencyID    uint32
	Kind        MessageKind
}

type Request struct {
	Header  RequestHeader
	Payload []byte
}

type Ack struct {
	BetNumber uint32
}

func NewBetRequest(agencyId uint32, m Marshaler) Request {
	payload, count := m.MarshalBet()
	payloadSize := len(payload)
	header := RequestHeader{
		PayloadSize: uint32(payloadSize),
		Count:       uint32(count),
		AgencyID:    agencyId,
		Kind:        PostBet,
	}
	return Request{Header: header, Payload: payload}
}

func (r Request) bytes() []byte {
	var data bytes.Buffer
	data.WriteByte(uint8(r.Header.Kind))
	buf := make([]byte, 0, 12)
	buf = binary.LittleEndian.AppendUint32(buf, r.Header.PayloadSize)
	buf = binary.LittleEndian.AppendUint32(buf, r.Header.Count)
	buf = binary.LittleEndian.AppendUint32(buf, r.Header.AgencyID)
	data.Write(buf)
	data.Write(r.Payload)
	return data.Bytes()
}

func EncodeRequest(w io.Writer, req Request) error {
	reqBytes := req.bytes()
	written := 0
	for written < len(reqBytes) {
		n, err := w.Write(reqBytes[written:])
		if err != nil {
			if !errors.Is(err, io.ErrShortWrite) {
				return errors.Wrap(err, "can't encode request")
			}
		}
		written += n
	}
	return nil
}

func DecodeResponse(r io.Reader) (Ack, error) {
	// Make constant
	buf := make([]byte, ResponseSize)
	read := 0
	for read < ResponseSize {
		n, err := r.Read(buf[read:])
		if err != nil {
			if errors.Is(err, io.EOF) && read < ResponseSize {
				return Ack{}, errors.Errorf("Unexpected EOF")
			}
			return Ack{}, errors.Wrap(err, "can't decode response")
		}
		read += n
	}
	betNumber := binary.LittleEndian.Uint32(buf[:])
	return Ack{
		BetNumber: betNumber,
	}, nil
}
