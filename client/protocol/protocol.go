package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
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
	_                                     = iota
	MessageBet                MessageKind = 0
	MessageAcknowledge                    = 1
	MessageDone                           = 2
	MessageWinners                        = 3
	MessageWinnersUnavailable             = 4
	MessageWinnersList                    = 5
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

type Done struct {
	ID uint32
}

type Winners struct {
	ID uint32
}

func (w Winners) Marshal() []byte {
	buf := make([]byte, 5, 5)
	buf[0] = MessageWinners
	binary.LittleEndian.PutUint32(buf[1:5], w.ID)
	return buf
}

func (d Done) Marshal() []byte {
	buf := make([]byte, 5, 5)
	buf[0] = MessageDone
	binary.LittleEndian.PutUint32(buf[1:5], d.ID)
	return buf
}

type Response interface {
	isResponse()
}

type Acknowledge struct {
	BetCount   uint32
	BetNumbers []uint32
}

func (a Acknowledge) isResponse() {}

type WinnersUnavailable struct{}

func (w WinnersUnavailable) isResponse() {}

type WinnersList struct {
	WinnerCount uint32
	DNIS        []string
}

func (w WinnersList) isResponse() {}

func EncodeRequest(w io.Writer, req Request) error {
	reqBytes := req.Marshal()
	if err := writeExact(w, reqBytes); err != nil {
		return errors.Wrap(err, "couldn't encode request")
	}
	return nil
}

func DecodeResponse(r io.Reader) (Response, error) {
	kindByte := make([]byte, 1)
	if err := readExact(r, kindByte); err != nil {
		return Acknowledge{}, errors.Wrap(err, "couldn't read kind byte")
	}

	kind := MessageKind(kindByte[0])
	switch kind {
	case MessageAcknowledge:
		countBytes := make([]byte, 4)
		if err := readExact(r, countBytes); err != nil {
			return nil, errors.Wrap(err, "couldn't read count")
		}
		betCount := binary.LittleEndian.Uint32(countBytes[0:4])
		nBytes := int(betCount) * int(unsafe.Sizeof(betCount))
		betNumbersBytes := make([]byte, nBytes)
		if err := readExact(r, betNumbersBytes); err != nil {
			return nil, errors.Wrap(err, "couldn't read bet numbers")
		}

		br := bytes.NewReader(betNumbersBytes)
		betNumbers := make([]uint32, betCount)
		if err := binary.Read(br, binary.LittleEndian, betNumbers); err != nil {
			return nil, errors.Wrap(err, "couldn't parse bet numbers")
		}

		return Acknowledge{
			BetCount:   betCount,
			BetNumbers: betNumbers,
		}, nil
	case MessageWinnersUnavailable:
		return WinnersUnavailable{}, nil
	case MessageWinnersList:
		buf := make([]byte, 8)
		if err := readExact(r, buf); err != nil {
			return nil, errors.Wrap(err, "couldn't read winners")
		}
		winners := binary.LittleEndian.Uint32(buf[:4])
		payload_size := binary.LittleEndian.Uint32(buf[4:])
		payload := make([]byte, payload_size)
		if err := readExact(r, payload); err != nil {
			return nil, errors.Wrap(err, "couldn't read winners")
		}
		dnis := strings.Split(string(payload), ",")
		return WinnersList{winners, dnis}, nil
	default:
		err := errors.Errorf("unknown message type: %d", kind)
		return nil, err
	}
}
