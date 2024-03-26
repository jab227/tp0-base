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

	t.Run("make done request", func(t *testing.T) {
		req := protocol.Done{2}
		var got bytes.Buffer
		err := protocol.EncodeRequest(&got, req)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		want := []byte{protocol.MessageDone, 2, 0, 0, 0}
		if !reflect.DeepEqual(got.Bytes(), want) {
			t.Errorf("got %v, want %v", got.Bytes(), want)
		}

	})

	t.Run("request winners", func(t *testing.T) {
		req := protocol.Winners{2}
		var got bytes.Buffer
		err := protocol.EncodeRequest(&got, req)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		want := []byte{protocol.MessageWinners, 2, 0, 0, 0}
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

func TestDecodeResponse(t *testing.T) {
	t.Run("decode acknowledge", func(t *testing.T) {
		const (
			betNumberA = 8
			betNumberB = 12
		)
		want := protocol.Acknowledge{
			BetCount:   2,
			BetNumbers: []uint32{betNumberA, betNumberB},
		}

		var buf bytes.Buffer
		buf.WriteByte(protocol.MessageAcknowledge)
		bs := make([]byte, 12, 12)
		binary.LittleEndian.PutUint32(bs[:4], want.BetCount)
		binary.LittleEndian.PutUint32(bs[4:8], want.BetNumbers[0])
		binary.LittleEndian.PutUint32(bs[8:12], want.BetNumbers[1])
		buf.Write(bs)

		res, err := protocol.DecodeResponse(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		got, ok := res.(protocol.Acknowledge)
		if !ok {
			t.Errorf("Expected Acknowledge")
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("decode winners_unavailable", func(t *testing.T) {
		const (
			betNumberA = 8
			betNumberB = 12
		)

		var buf bytes.Buffer
		buf.WriteByte(protocol.MessageWinnersUnavailable)

		res, err := protocol.DecodeResponse(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		_, ok := res.(protocol.WinnersUnavailable)
		if !ok {
			t.Errorf("Expected winners_unavailable")
		}
	})

	t.Run("decode winners_list", func(t *testing.T) {
		const WinnerCount = 42

		want := protocol.WinnersList{
			WinnerCount: WinnerCount,
			DNIS:        []string{"1856", "1812"},
		}

		var buf bytes.Buffer
		buf.WriteByte(protocol.MessageWinnersList)
		bs := make([]byte, 4, 4)
		binary.LittleEndian.PutUint32(bs, WinnerCount)
		buf.Write(bs)
		payload := []byte("1856,1812")
		binary.LittleEndian.PutUint32(bs, uint32(len(payload)))
		buf.Write(bs)		
		buf.Write(payload)

		res, err := protocol.DecodeResponse(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		got, ok := res.(protocol.WinnersList)
		if !ok {
			t.Errorf("Expected winners_list")
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
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
	want := protocol.Acknowledge{
		BetCount:   2,
		BetNumbers: []uint32{betNumberA, betNumberB},
	}

	var buf bytes.Buffer
	buf.WriteByte(protocol.MessageAcknowledge)
	bs := make([]byte, 12, 12)
	binary.LittleEndian.PutUint32(bs[:4], want.BetCount)
	binary.LittleEndian.PutUint32(bs[4:8], want.BetNumbers[0])
	binary.LittleEndian.PutUint32(bs[8:12], want.BetNumbers[1])
	buf.Write(bs)

	r := &ShortReader{buf: &buf}
	res, err := protocol.DecodeResponse(r)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	got, ok := res.(protocol.Acknowledge)
	if !ok {
		t.Errorf("Expected Acknowledge")
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if r.calls != expectedCalls {
		t.Errorf("got %v, want %v", r.calls, expectedCalls)
	}
}
