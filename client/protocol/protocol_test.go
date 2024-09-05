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

	const agencyID = 42

	bet, _ := agency.NewBet(bettor)
	req := protocol.NewBetRequest(agencyID, bet)
	payload := []byte(fmt.Sprintf("%s,%s,%s,%s,%s",
		bettor.Name,
		bettor.Surname,
		bettor.DNI,
		bettor.Birthdate,
		bettor.BetNumber))

	var got bytes.Buffer
	err := protocol.EncodeRequest(&got, req)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	var want bytes.Buffer

	want.WriteByte(uint8(protocol.Bet))

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
		betNumber = 8
	)
	want := protocol.BetAcknowledge{
		BetNumber: betNumber,
	}

	var data bytes.Buffer
	buf := make([]byte, protocol.BetAcknowledgeSize)
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

	const agencyID = 42
	bet, _ := agency.NewBet(bettor)
	req := protocol.NewBetRequest(agencyID, bet)

	var got bytes.Buffer
	w := ShortWriter{buf: &got}

	err := protocol.EncodeRequest(&w, req)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	var want bytes.Buffer
	payload := []byte(fmt.Sprintf("%s,%s,%s,%s,%s",
		bettor.Name,
		bettor.Surname,
		bettor.DNI,
		bettor.Birthdate,
		bettor.BetNumber))
	want.WriteByte(uint8(protocol.Bet))

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(payload)))
	want.Write(buf)

	binary.LittleEndian.PutUint32(buf, uint32(agencyID))
	want.Write(buf)

	want.Write(payload)

	if !reflect.DeepEqual(got.Bytes(), want.Bytes()) {
		t.Errorf("got %v, want %v", got.Bytes(), want.Bytes())
	}

	if w.calls != int(protocol.RequestHeaderSize)+len(payload) {
		fmt.Printf("%v\n", reflect.TypeOf(protocol.RequestHeader{}).Size())
		t.Errorf("got %v, want %v", w.calls, int(protocol.RequestHeaderSize)+len(payload))
	}
}

func TestShortRead(t *testing.T) {
	const (
		betNumber = 8
	)
	want := protocol.BetAcknowledge{
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

	if r.calls != int(protocol.BetAcknowledgeSize) {
		t.Errorf("got %v, want %v", r.calls, int(protocol.BetAcknowledgeSize))
	}
}
