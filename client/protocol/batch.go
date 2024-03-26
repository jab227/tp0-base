package protocol

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/pkg/errors"
)

const KB = (1 << 10)
const MaxBatchByteSize = 8 * KB

type BetBatcher struct {
	bets       []agency.Bet
	scanner    *bufio.Scanner
	notBatched int
	batchSize  int
	written    int
}

func NewBatcher(r io.Reader, batchSize int) *BetBatcher {
	return &BetBatcher{
		batchSize: batchSize,
		scanner:   bufio.NewScanner(r),
	}
}

func (b *BetBatcher) Written() int {
	return b.written
}

func (b *BetBatcher) MarshalBet() ([]byte, int) {
	requestSize := RequestHeaderSize
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
	b.written += i
	payload := buf.Bytes()
	return payload[:len(payload)-1], i
}

func (b *BetBatcher) ReadBatch() (bool, error) {
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
