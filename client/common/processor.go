package common

import (
	"bufio"
	"fmt"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/batch"
	"github.com/pkg/errors"
	"io"
	"strings"
	"sync"
)

type BetResult struct {
	Bet agency.Bet
	Err error
}

func parseBetFromLine(line string) (agency.Bet, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 5 {
		err := fmt.Errorf("not enough fields %d", len(fields))
		return agency.Bet{}, err
	}

	b, err := agency.NewBet(agency.Bettor{
		Name:      fields[0],
		Surname:   fields[1],
		DNI:       fields[2],
		Birthdate: fields[3],
		BetNumber: fields[4],
	})
	if err != nil {
		err := errors.Wrap(err, "couldn't parse bet")
		return agency.Bet{}, err
	}
	return b, nil
}

func BetScannerRun(r io.ReadCloser, wg *sync.WaitGroup, done <-chan struct{}) <-chan BetResult {
	bets := make(chan BetResult, 1)
	go func() {
		defer r.Close()
		defer wg.Done()
		defer close(bets)
		var lineNumber int = 1
		scanner := bufio.NewScanner(r)
		for {
			select {
			case <-done:
				fmt.Println("done scanner")
				return
			default:
				if !scanner.Scan() {
					return
				}
				l := scanner.Text()
				if l == "" {
					return
				}
				fmt.Println(l)
				bet, err := parseBetFromLine(l)
				if err != nil {
					err = errors.Wrapf(err, "line %d", lineNumber)
					bets <- BetResult{Err: err}
					return
				}
				bets <- BetResult{Bet: bet}
				lineNumber += 1
			}
			if err := scanner.Err(); err != nil {
				bets <- BetResult{Err: err}
				return
			}
		}
	}()
	return bets
}

type BatchResult struct {
	Chunk batch.Chunk
	Err   error
}

func BatchProcessor(
	b *batch.Batcher,
	wg *sync.WaitGroup,
	bets <-chan BetResult,
	done <-chan struct{},
) <-chan BatchResult {
	resultCh := make(chan BatchResult, 1)
	go func() {
		defer wg.Done()
		defer close(resultCh)
		for {
			select {
			case <-done:
				fmt.Println("done batcher")
				return
			case r, more := <-bets:
				if !more {
					return
				}
				if r.Err != nil {
					err := errors.Wrap(r.Err, "couldn't batch bet")
					resultCh <- BatchResult{Err: err}
					return
				}

				bet := r.Bet
				b.Push(bet)
				chunk, ok := b.Next()
				if !ok {
					continue
				}
				resultCh <- BatchResult{Chunk: chunk}
			}
		}
	}()
	return resultCh
}
