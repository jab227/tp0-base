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
	_                              = iota
	MessageBet         MessageKind = 0
	MessageAcknowledge             = 1
	MessageDone                    = 2
)

type Request interface {
	Marshal() []byte
}

type Bet struct {
	PayloadSize uint32
	Count       uint32
	AgencyID    uint32
	Payload     []byte
}

func (b Bet) Marshal() []byte {
	var data bytes.Buffer

	data.WriteByte(uint8(MessageBet))

	buf := make([]byte, 12, 12)

	binary.LittleEndian.PutUint32(buf[:4], b.PayloadSize)
	binary.LittleEndian.PutUint32(buf[4:8], b.Count)
	binary.LittleEndian.PutUint32(buf[8:12], b.AgencyID)

	data.Write(buf)
	data.Write(b.Payload)
	return data.Bytes()
}

type Done struct{}

func (d Done) Marshal() []byte {
	buf := make([]byte, 1, 1)
	buf[0] = MessageDone
	return buf
}

type Response struct {
	Kind       MessageKind
	BetCount   uint32
	BetNumbers []uint32
}

func EncodeRequest(w io.Writer, req Request) error {
	reqBytes := req.Marshal()
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
		return Response{}, errors.Wrap(err, "couldn't read kind byte")
	}

	kind := MessageKind(kindByte[0])
	if kind != MessageAcknowledge {
		err := errors.Errorf("unknown message type: %d", kind)
		return Response{}, err
	}

	countBytes := make([]byte, 4)
	if err := readExact(r, countBytes); err != nil {
		return Response{}, errors.Wrap(err, "couldn't read count")
	}
	betCount := binary.LittleEndian.Uint32(countBytes[0:4])
	nBytes := int(betCount) * int(unsafe.Sizeof(betCount))
	betNumbersBytes := make([]byte, nBytes)
	if err := readExact(r, betNumbersBytes); err != nil {
		return Response{}, errors.Wrap(err, "couldn't read bet numbers")
	}

	br := bytes.NewReader(betNumbersBytes)
	betNumbers := make([]uint32, betCount)
	if err := binary.Read(br, binary.LittleEndian, betNumbers); err != nil {
		return Response{}, errors.Wrap(err, "couldn't parse bet numbers")
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
