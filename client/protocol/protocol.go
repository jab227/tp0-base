package protocol

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// Byte sizes
const (
	ResponseSize      = 8
	RequestHeaderSize = 9
)

type Marshaler interface {
	MarshalBet() []byte
}

type MessageKind uint8

const (
	_                   = iota
	PostBet MessageKind = 0
)

type RequestHeader struct {
	PayloadSize uint32
	AgencyID    uint32
	Kind        MessageKind
}

type Request struct {
	Header  RequestHeader
	Payload []byte
}

type Ack struct {
	AgencyID  uint32
	BetNumber uint32
}

func NewBetRequest(agencyId uint32, m Marshaler) Request {
	payload := m.MarshalBet()
	payloadSize := len(payload)
	header := RequestHeader{
		PayloadSize: uint32(payloadSize),
		AgencyID:    agencyId,
		Kind:        PostBet,
	}
	return Request{Header: header, Payload: payload}
}

func (r Request) bytes() []byte {
	var data bytes.Buffer
	data.WriteByte(uint8(r.Header.Kind))

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(r.Header.PayloadSize))
	data.Write(buf)
	binary.LittleEndian.PutUint32(buf, uint32(r.Header.AgencyID))
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

			}
			return Ack{}, errors.Wrap(err, "can't decode response")
		}
		read += n
	}
	agencyId := binary.LittleEndian.Uint32(buf[:4])
	betNumber := binary.LittleEndian.Uint32(buf[4:])
	return Ack{
		AgencyID:  agencyId,
		BetNumber: betNumber,
	}, nil
}
