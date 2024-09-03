package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
	"github.com/pkg/errors"
	"io"
)

// Byte sizes
var (
	BetAcknowledgeSize = utils.PackedSizeOf(BetAcknowledge{})
	HeaderSize         = utils.PackedSizeOf(Header{})
)

type Marshaler interface {
	MarshalPayload() []byte
}

type MessageKind uint8

const (
	Bet      MessageKind = 0
	BetBatch             = 1
	BetBatchStop
)

type Header struct {
	PayloadSize uint32
	AgencyID    uint32
	Kind        MessageKind
}

type BetRequest struct {
	h       Header
	payload []byte
}

type BetAcknowledge struct {
	BetNumber uint32
}

func NewBetRequest(kind MessageKind, agencyId uint32, m Marshaler) BetRequest {
	var payload []byte
	if m != nil {
		payload = m.MarshalPayload()
	}
	payloadSize := len(payload)
	h := Header{
		PayloadSize: uint32(payloadSize),
		AgencyID:    agencyId,
		Kind:        kind,
	}
	return BetRequest{h: h, payload: payload}
}

func (r BetRequest) toBytes() []byte {
	var data bytes.Buffer
	data.WriteByte(uint8(r.h.Kind))

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, r.h.PayloadSize)
	data.Write(buf)
	binary.LittleEndian.PutUint32(buf, r.h.AgencyID)
	data.Write(buf)
	data.Write(r.payload)
	return data.Bytes()
}

func EncodeRequest(w io.Writer, req BetRequest) error {
	reqBytes := req.toBytes()
	return utils.WriteAll(w, reqBytes)
}

func DecodeResponse(r io.Reader) (BetAcknowledge, error) {
	// Make constant
	buf := make([]byte, BetAcknowledgeSize)
	if err := utils.ReadAtLeast(r, buf); err != nil {
		err = errors.Wrap(err, "couldn't decode response")
		return BetAcknowledge{}, err
	}
	betNumber := binary.LittleEndian.Uint32(buf[:])
	return BetAcknowledge{
		BetNumber: betNumber,
	}, nil
}
