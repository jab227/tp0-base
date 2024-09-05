package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
	"io"
)

type Marshaler interface {
	MarshalPayload() []byte
}

type RequestKind uint8

const (
	_             RequestKind = iota
	Bet                       = 0
	BetBatch                  = 1
	BetBatchStop              = 2
	BetGetWinners             = 3
)

var RequestHeaderSize = utils.PackedSizeOf(RequestHeader{})

type RequestHeader struct {
	PayloadSize uint32
	AgencyID    uint32
	Kind        RequestKind
}

type BetRequest struct {
	h       RequestHeader
	payload []byte
}

func NewBetRequest(kind RequestKind, agencyId uint32, m Marshaler) BetRequest {
	var payload []byte
	if m != nil {
		payload = m.MarshalPayload()
	}
	payloadSize := len(payload)
	h := RequestHeader{
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
