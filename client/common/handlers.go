package common

import (
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const BackoffExp = 2

var ErrUnexpectedResponse = errors.New("unexpected response")

type HandlerResponse struct {
	Response protocol.Response
	Err      error
}

type ProtocolHandlerContext struct {
	Results       chan<- HandlerResponse
	Done          chan<- struct{}
	Requests      <-chan protocol.Request
	Timeout       time.Duration
	Backoff       time.Duration
	MaxRetries    int
	ServerAddress string
	ID            uint32
}

type BatchHandlerContext struct {
	ID        uint32
	Reader    io.Reader
	BatchSize int
	Requests  chan<- protocol.Request
}

func HandleSignals(client *Client, ch <-chan os.Signal, done chan struct{}) {
	client.done = done
	go func(id int) {
		signal := <-ch
		log.Infof("action: signal | result: success | client_id: %v | msg: received %s", id, signal)
		done <- struct{}{}
	}(int(client.config.ID))

}

func HandleProtocol(ctx *ProtocolHandlerContext, conn net.Conn) {
	for req := range ctx.Requests {
		conn.SetWriteDeadline(time.Now().Add(ctx.Timeout))
		if err := protocol.EncodeRequest(conn, req); err != nil {
			ctx.Results <- HandlerResponse{Err: err}
			return
		}
		conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
		res, err := protocol.DecodeResponse(conn)
		if err != nil {
			ctx.Results <- HandlerResponse{Err: err}
			return
		}
		ack, ok := res.(protocol.Acknowledge)
		var serverResponse HandlerResponse
		if !ok {
			serverResponse = HandlerResponse{Err: errors.Wrap(ErrUnexpectedResponse, "expected acknowledge")}

		} else {
			serverResponse = HandlerResponse{Response: ack}
		}
		ctx.Results <- serverResponse
	}

	req := protocol.Done{ID: ctx.ID}
	conn.SetWriteDeadline(time.Now().Add(ctx.Timeout))
	if err := protocol.EncodeRequest(conn, req); err != nil {
		ctx.Results <- HandlerResponse{Err: err}
		return
	}
	log.Debug("Done sent")
	GetWinners(ctx, conn)

	return
}

func GetWinners(ctx *ProtocolHandlerContext, conn net.Conn) {
	retries := 0
	backoff := ctx.Backoff

	for retries < ctx.MaxRetries {
		req := protocol.Winners{ID: ctx.ID}
		conn.SetWriteDeadline(time.Now().Add(ctx.Timeout))
		if err := protocol.EncodeRequest(conn, req); err != nil {
			ctx.Results <- HandlerResponse{Err: err}
			return
		}

		conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
		res, err := protocol.DecodeResponse(conn)
		if err != nil {
			ctx.Results <- HandlerResponse{Err: err}
			return
		}

		switch r := res.(type) {
		case protocol.WinnersUnavailable:
			time.Sleep(backoff)
			retries++
			backoff *= BackoffExp
			backoff += time.Duration(rand.Int63n(100))
		case protocol.WinnersList:
			ctx.Results <- HandlerResponse{Response: r}
			return
		default:
			retries++
			err := errors.Wrap(ErrUnexpectedResponse, "expected winners_list or winners_unavailable messages")
			ctx.Results <- HandlerResponse{Err: err}
			return
		}
	}
	err := errors.Errorf("couldn't get winners: max retry attemps reached %d", ctx.MaxRetries)
	ctx.Results <- HandlerResponse{Err: err}
}

func HandleBetsBatching(ctx *BatchHandlerContext) {
	defer close(ctx.Requests)
	batcher := protocol.NewBatcher(ctx.Reader, ctx.BatchSize)

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
