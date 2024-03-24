package protocol

import (
	"bytes"
	"encoding/binary"
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
	SendBets    MessageKind = 0
	Acknowledge             = 1
	EndBets                 = 2
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

type Response struct {
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
		Kind:        SendBets,
	}
	return Request{Header: header, Payload: payload}
}

func (r Request) bytes() []byte {
	var data bytes.Buffer
	data.WriteByte(uint8(r.Header.Kind))
	buf := make([]byte, 12, 12)
	if r.Header.Kind == SendBets {
		binary.LittleEndian.PutUint32(buf[:4], r.Header.PayloadSize)
		binary.LittleEndian.PutUint32(buf[4:8], r.Header.Count)
		binary.LittleEndian.PutUint32(buf[8:12], r.Header.AgencyID)
	}
	data.Write(buf)
	data.Write(r.Payload)
	return data.Bytes()
}

func EncodeRequest(w io.Writer, req Request) error {
	reqBytes := req.bytes()
	if err := writeExact(w, reqBytes); err != nil {
		return errors.Wrap(err, "couldn't encode request")
	}
	return nil
}

func writeExact(w io.Writer, p []byte) error {
	written := 0
	for written < len(p) {
		n, err := w.Write(p[written:])
		if err != nil {
			if !errors.Is(err, io.ErrShortWrite) {
				return err
			}
		}
		written += n
	}
	return nil
}

func DecodeResponse(r io.Reader) (Response, error) {
	kindByte := make([]byte, 1)
	if err := readExact(r, kindByte); err != nil {
		return Response{}, errors.Wrap(err, "couldn't decode response: read ack header")
	}

	kind := MessageKind(kindByte[0])
	if kind != Acknowledge {
		return Response{}, errors.Errorf("unknown message type: %d", kind)
	}

	countBytes := make([]byte, 4)
	if err := readExact(r, countBytes); err != nil {
		return Response{}, errors.Wrap(err, "couldn't decode response: count bytes")
	}
	betCount := binary.LittleEndian.Uint32(countBytes[0:4])
	nBytes := int(betCount) * int(unsafe.Sizeof(betCount))
	betNumbersBytes := make([]byte, nBytes)
	if err := readExact(r, betNumbersBytes); err != nil {
		return Response{}, errors.Wrap(err, "couldn't decode response: read bet numbers")
	}

	br := bytes.NewReader(betNumbersBytes)
	betNumbers := make([]uint32, betCount)
	if err := binary.Read(br, binary.LittleEndian, betNumbers); err != nil {
		return Response{}, errors.Wrap(err, "couldn't decode response: parse bet numbers")
	}

	return Response{
		Kind:       kind,
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
