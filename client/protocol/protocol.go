package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/pkg/errors"
)

// Byte sizes
const (
	RequestHeaderSize = 13
)

type Marshaler interface {
	MarshalBet() ([]byte, int)
}

type MessageKind uint8

const (
	_                       = iota
	PostBet     MessageKind = 0
	Acknowledge             = 1
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
	Kind       MessageKind
	BetCount   uint32
	BetNumbers []uint32
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
	ackHeader := make([]byte, 5)
	if err := readExact(r, ackHeader); err != nil {
		return Ack{}, errors.Wrap(err, "couldn't decode response")
	}

	kind := ackHeader[0]
	if kind != Acknowledge {
		return Ack{}, errors.Errorf("unknown message type: %d", kind)
	}

	betCount := binary.LittleEndian.Uint32(ackHeader[1:5])
	nBytes := int(betCount) * int(unsafe.Sizeof(betCount))
	betNumbersBytes := make([]byte, nBytes)
	if err := readExact(r, betNumbersBytes); err != nil {
		return Ack{}, errors.Wrap(err, "couldn't decode response")
	}

	br := bytes.NewReader(betNumbersBytes)
	fmt.Println(betNumbersBytes)
	betNumbers := make([]uint32, betCount)
	if err := binary.Read(br, binary.LittleEndian, betNumbers); err != nil {
		return Ack{}, errors.Wrap(err, "couldn't decode response")
	}

	return Ack{
		Kind:       MessageKind(kind),
		BetCount:   betCount,
		BetNumbers: betNumbers,
	}, nil
}

func readExact(r io.Reader, p []byte) error {
	read := 0
	size := len(p)
	for read < size {
		n, err := r.Read(p[read:])
		if err != nil {
			return err
		}
		read += n
	}
	return nil
}
