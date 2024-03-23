package protocol_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

func TestPostBetRequest(t *testing.T) {
	bettor := agency.Bettor{
		Name:      "Julio",
		Surname:   "Cortazar",
		DNI:       "52820003",
		Birthdate: "1999-03-17",
		BetNumber: "7574",
	}

	const (
		agencyID = 42
		betCount = 1
	)

	bet, _ := agency.NewBet(bettor)

	req := protocol.NewBetRequest(agencyID, bet)
	payload := []byte(fmt.Sprintf("%s,%s,%s,%s,%s",
		bettor.Name,
		bettor.Surname,
		bettor.DNI,
		bettor.Birthdate,
		bettor.BetNumber))

	expectedHeader := protocol.RequestHeader{
		Kind:        protocol.PostBet,
		AgencyID:    agencyID,
		Count:       1,
		PayloadSize: uint32(len(payload)),
	}
	expectedReq := protocol.Request{
		Header:  expectedHeader,
		Payload: payload,
	}

	if !reflect.DeepEqual(req, expectedReq) {
		t.Errorf("got %v, want %v", req, expectedReq)
	}

	var got bytes.Buffer
	err := protocol.EncodeRequest(&got, req)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	want := getRequestBytes(payload, agencyID, betCount)

	if !reflect.DeepEqual(got.Bytes(), want) {
		t.Errorf("got %v, want %v", got.Bytes(), want)
	}

}

func getRequestBytes(payload []byte, agencyID uint32, betCount uint32) []byte {
	var want bytes.Buffer

	want.WriteByte(uint8(protocol.PostBet))

	buf := make([]byte, 0, 12)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(payload)))
	buf = binary.LittleEndian.AppendUint32(buf, betCount)
	buf = binary.LittleEndian.AppendUint32(buf, agencyID)
	want.Write(buf)

	want.Write(payload)
	return want.Bytes()
}

func TestDecodeServerResponse(t *testing.T) {
	const (
		betNumber = 8
	)
	want := protocol.Ack{
		BetNumber: betNumber,
	}

	var data bytes.Buffer
	buf := make([]byte, protocol.ResponseSize)
	binary.LittleEndian.PutUint32(buf, betNumber)
	data.Write(buf)

	got, err := protocol.DecodeResponse(&data)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

type ShortWriter struct {
	buf   *bytes.Buffer
	calls int
}

func (s *ShortWriter) Write(b []byte) (int, error) {
	s.buf.Write(b[:1])
	s.calls++
	if len(b) != 1 {
		return 1, io.ErrShortWrite
	}
	return 1, nil
}

type ShortReader struct {
	buf   *bytes.Buffer
	calls int
}

func (s *ShortReader) Read(b []byte) (int, error) {
	s.calls++
	return s.buf.Read(b[:1])
}

func TestShortWrite(t *testing.T) {
	bettor := agency.Bettor{
		Name:      "Julio",
		Surname:   "Cortazar",
		DNI:       "52820003",
		Birthdate: "1999-03-17",
		BetNumber: "7574",
	}

	const (
		agencyID = 42
		betCount = 1
	)

	bet, _ := agency.NewBet(bettor)
	req := protocol.NewBetRequest(agencyID, bet)
	payload := []byte(fmt.Sprintf("%s,%s,%s,%s,%s",
		bettor.Name,
		bettor.Surname,
		bettor.DNI,
		bettor.Birthdate,
		bettor.BetNumber))

	var got bytes.Buffer
	w := ShortWriter{buf: &got}

	err := protocol.EncodeRequest(&w, req)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	want := getRequestBytes(payload, agencyID, betCount)

	if !reflect.DeepEqual(got.Bytes(), want) {
		t.Errorf("got %v, want %v", got.Bytes(), want)
	}

	if w.calls != protocol.RequestHeaderSize+len(payload) {
		t.Errorf("got %v, want %v", w.calls, protocol.RequestHeaderSize+len(payload))
	}
}

func TestShortRead(t *testing.T) {
	const (
		betNumber = 8
	)
	want := protocol.Ack{
		BetNumber: betNumber,
	}

	var data bytes.Buffer
	r := ShortReader{&data, 0}
	buf := make([]byte, 4)

	binary.LittleEndian.PutUint32(buf, betNumber)
	data.Write(buf)

	got, err := protocol.DecodeResponse(&r)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if r.calls != protocol.ResponseSize {
		t.Errorf("got %v, want %v", r.calls, protocol.ResponseSize)
	}
}
