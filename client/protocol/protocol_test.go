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

func TestEncodeRequest(t *testing.T) {
	t.Run("make a bet", func(t *testing.T) {
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

		expectedPayload := []byte(fmt.Sprintf("%s,%s,%s,%s,%s",
			bettor.Name,
			bettor.Surname,
			bettor.DNI,
			bettor.Birthdate,
			bettor.BetNumber))

		payload, count := bet.MarshalBet()

		if !reflect.DeepEqual(payload, expectedPayload) {
			t.Errorf("got %v, want %v", payload, expectedPayload)
		}

		if count != betCount {
			t.Errorf("got %v, want %v", count, betCount)
		}

		req := protocol.Bet{
			PayloadSize: uint32(len(payload)),
			Count:       uint32(count),
			AgencyID:    agencyID,
			Payload:     payload,
		}

		var got bytes.Buffer
		err := protocol.EncodeRequest(&got, req)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		want := getRequestBytes(expectedPayload, agencyID, betCount)

		if !reflect.DeepEqual(got.Bytes(), want) {
			t.Errorf("got %v, want %v", got.Bytes(), want)
		}
	})

	t.Run("make multiple bets", func(t *testing.T) {
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

		expectedPayload := []byte(fmt.Sprintf("%s,%s,%s,%s,%s",
			bettor.Name,
			bettor.Surname,
			bettor.DNI,
			bettor.Birthdate,
			bettor.BetNumber))

		payload, count := bet.MarshalBet()

		if !reflect.DeepEqual(payload, expectedPayload) {
			t.Errorf("got %v, want %v", payload, expectedPayload)
		}

		if count != betCount {
			t.Errorf("got %v, want %v", count, betCount)
		}

		req := protocol.Bet{
			PayloadSize: uint32(len(payload)),
			Count:       uint32(count),
			AgencyID:    agencyID,
			Payload:     payload,
		}

		var got bytes.Buffer
		err := protocol.EncodeRequest(&got, req)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		want := getRequestBytes(expectedPayload, agencyID, betCount)

		if !reflect.DeepEqual(got.Bytes(), want) {
			t.Errorf("got %v, want %v", got.Bytes(), want)
		}
	})

	t.Run("make done request", func(t *testing.T) {
		req := protocol.Done{}
		var got bytes.Buffer
		err := protocol.EncodeRequest(&got, req)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		want := []byte{protocol.MessageDone}
		if !reflect.DeepEqual(got.Bytes(), want) {
			t.Errorf("got %v, want %v", got.Bytes(), want)
		}

	})

}

func getRequestBytes(payload []byte, agencyID uint32, betCount uint32) []byte {
	var want bytes.Buffer

	want.WriteByte(uint8(protocol.MessageBet))

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
		betNumberA = 8
		betNumberB = 12
	)
	want := protocol.Response{
		Kind:       protocol.MessageAcknowledge,
		BetCount:   2,
		BetNumbers: []uint32{betNumberA, betNumberB},
	}

	var buf bytes.Buffer
	buf.WriteByte(protocol.MessageAcknowledge)
	bs := make([]byte, 0, 12)
	bs = binary.LittleEndian.AppendUint32(bs, want.BetCount)
	bs = binary.LittleEndian.AppendUint32(bs, want.BetNumbers[0])
	bs = binary.LittleEndian.AppendUint32(bs, want.BetNumbers[1])
	buf.Write(bs)
	got, err := protocol.DecodeResponse(&buf)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
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
	payload, count := bet.MarshalBet()
	req := protocol.Bet{
		PayloadSize: uint32(len(payload)),
		Count:       uint32(count),
		AgencyID:    agencyID,
		Payload:     payload,
	}

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
		betNumberA    = 8
		betNumberB    = 12
		expectedCalls = 13 // Sizeof Ack message in bytes
	)
	want := protocol.Response{
		Kind:       protocol.MessageAcknowledge,
		BetCount:   2,
		BetNumbers: []uint32{betNumberA, betNumberB},
	}

	var buf bytes.Buffer
	buf.WriteByte(protocol.MessageAcknowledge)
	bs := make([]byte, 0, 12)
	bs = binary.LittleEndian.AppendUint32(bs, want.BetCount)
	bs = binary.LittleEndian.AppendUint32(bs, want.BetNumbers[0])
	bs = binary.LittleEndian.AppendUint32(bs, want.BetNumbers[1])
	buf.Write(bs)

	r := &ShortReader{buf: &buf}
	got, err := protocol.DecodeResponse(r)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if r.calls != expectedCalls {
		t.Errorf("got %v, want %v", r.calls, expectedCalls)
	}
}
