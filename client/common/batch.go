package common

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const KB = (1 << 10)
const MaxBatchByteSize = 8 * KB

type Batcher struct {
	bets       []agency.Bet
	scanner    *bufio.Scanner
	notBatched int
	batchSize  int
}

func NewBatcher(r io.Reader, batchSize int) *Batcher {
	return &Batcher{
		batchSize: batchSize,
		scanner:   bufio.NewScanner(r),
	}
}

func (b *Batcher) MarshalBet() ([]byte, int) {
	requestSize := protocol.RequestHeaderSize
	var buf bytes.Buffer
	i := 0
	for _, bet := range b.bets {
		if i >= b.batchSize {
			break
		}
		betData, _ := bet.MarshalBet()
		if requestSize+len(betData) > MaxBatchByteSize {
			break
		}
		buf.Write(betData)
		buf.WriteByte('\n')
		requestSize += (len(betData) + 1)
		i++
	}
	b.bets = b.bets[i:]
	b.notBatched -= i
	payload := buf.Bytes()
	return payload[:len(payload)-1], i
}

func (b *Batcher) ReadBatch() (bool, error) {
	lineNumber := 1
	read := 0
	for read < b.batchSize {
		if !b.scanner.Scan() {
			break
		}
		line := b.scanner.Text()
		fields := strings.Split(line, ",")
		bettor := agency.Bettor{
			Name:      fields[0],
			Surname:   fields[1],
			DNI:       fields[2],
			Birthdate: fields[3],
			BetNumber: fields[4],
		}

		bet, err := agency.NewBet(bettor)
		if err != nil {
			err = errors.Wrapf(err, "couldn't parse line %d", lineNumber)
			return false, err

		}
		b.bets = append(b.bets, bet)
		read++
		lineNumber++
	}
	b.notBatched += read
	return b.notBatched == 0, nil
}

type BatchContext struct {
	ID        uint32
	Reader    io.Reader
	BatchSize int
	Requests  chan<- protocol.Request
}

func HandleBatchs(ctx *BatchContext) {
	defer close(ctx.Requests)
	batcher := NewBatcher(ctx.Reader, ctx.BatchSize)

	for {
		finished, err := batcher.ReadBatch()
		if finished {
			break
		}
		if err != nil {
			// Assume that the previous batches were ok
			// Close the channel and don't read more bets
			log.Errorf("action: error_detected | result: success | client_id: %v | error: %s", ctx.ID, err.Error())
			break
		}
		payload, count := batcher.MarshalBet()
		req := protocol.Bet{
			PayloadSize: uint32(len(payload)),
			Count:       uint32(count),
			AgencyID:    ctx.ID,
			Payload:     payload,
		}
		ctx.Requests <- req
	}
}
