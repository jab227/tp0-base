package protocol_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

func TestPostBetRequest(t *testing.T) {

	const (
		name      = "Julio"
		surname   = "Cortazar"
		dni       = "52820003"
		birthdate = "1999-03-17"
		betNumber = "7574"
		agencyID  = 42
	)
	bet, _ := agency.NewBet(name, surname, dni, birthdate, betNumber)
	req := protocol.NewBetRequest(agencyID, bet)
	payload := []byte(fmt.Sprintf("%s,%s,%s,%s,%s", name, surname, dni, birthdate, betNumber))
	expected_header := protocol.RequestHeader{
		Kind:        protocol.PostBet,
		AgencyID:    agencyID,
		PayloadSize: uint32(len(payload)),
	}
	expectec_req := protocol.Request{
		Header:  expected_header,
		Payload: payload,
	}

	if !reflect.DeepEqual(req, expectec_req) {
		t.Errorf("got %v, want %v", req, expectec_req)
	}

	var got bytes.Buffer
	protocol.EncodeRequest(&got, req)

	var want bytes.Buffer

	want.WriteByte(uint8(protocol.PostBet))

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(payload)))
	want.Write(buf)

	binary.LittleEndian.PutUint32(buf, uint32(agencyID))
	want.Write(buf)

	want.Write(payload)

	if !reflect.DeepEqual(got.Bytes(), want.Bytes()) {
		t.Errorf("got %v, want %v", got.Bytes(), want.Bytes())
	}
}

func TestDecodeServerResponse(t *testing.T) {
	const (
		agencyID  = 42
		betNumber = 8
	)
	want := protocol.Ack{
		AgencyID:  agencyID,
		BetNumber: betNumber,
	}

	var data bytes.Buffer
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, agencyID)
	data.Write(buf)
	binary.LittleEndian.PutUint32(buf, betNumber)
	data.Write(buf)

	got := protocol.DecodeResponse(&data)
	
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
