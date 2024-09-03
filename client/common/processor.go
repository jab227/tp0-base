package common

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/batch"
	"github.com/pkg/errors"
)

type betResult struct {
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

type BatchResult struct {
	Chunk batch.Chunk
	Err   error
}

type BatchProcessor struct {
	done    <-chan struct{}
	bets    <-chan betResult
	rc      io.ReadCloser
	batcher *batch.Batcher
	wg      *sync.WaitGroup
}

type BatcherConfig struct {
	MaxCount int
	MaxSize  int
}

func NewBatchProcessor(
	rc io.ReadCloser,
	conf BatcherConfig,
	done <-chan struct{},
) *BatchProcessor {
	wg := new(sync.WaitGroup)
	batcher := batch.NewBatcher(conf.MaxCount, conf.MaxSize)
	return &BatchProcessor{
		batcher: batcher,
		wg:      wg,
		done:    done,
		rc:      rc,
	}
}

func (b *BatchProcessor) betScannerStart() <-chan betResult {
	betsCh := make(chan betResult, 1)
	go func() {
		defer b.rc.Close()
		defer b.wg.Done()
		defer close(betsCh)

		var lineNumber int = 1
		scanner := bufio.NewScanner(b.rc)
		for {
			select {
			case <-b.done:
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
				bet, err := parseBetFromLine(l)
				if err != nil {
					err = errors.Wrapf(err, "line %d", lineNumber)
					betsCh <- betResult{Err: err}
					return
				}
				betsCh <- betResult{Bet: bet}
				lineNumber += 1
			}
			if err := scanner.Err(); err != nil {
				betsCh <- betResult{Err: err}
				return
			}
		}
	}()
	return betsCh
}

func (b *BatchProcessor) batcherStart() <-chan BatchResult {
	resultCh := make(chan BatchResult, 1)
	go func() {
		defer b.wg.Done()
		defer close(resultCh)
		for {
			select {
			case <-b.done:
				fmt.Println("done batcher")
				return
			case r, more := <-b.bets:
				if !more {
					return
				}
				if r.Err != nil {
					err := errors.Wrap(r.Err, "couldn't batch bet")
					resultCh <- BatchResult{Err: err}
					return
				}

				bet := r.Bet
				b.batcher.Push(bet)
				chunk, ok := b.batcher.Next()
				if !ok {
					continue
				}
				log.Debug("looping batch")
				resultCh <- BatchResult{Chunk: chunk}
			}
		}
	}()
	return resultCh
}

func (b *BatchProcessor) Wait() {
	b.wg.Wait()
}

func (b *BatchProcessor) Run() <-chan BatchResult {
	b.wg.Add(2)
	b.bets = b.betScannerStart()
	return b.batcherStart()
}
