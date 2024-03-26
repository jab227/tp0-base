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

// ClientConfig Configuration used by the client
type ClientConfig struct {
	Timeout       time.Duration
	Backoff       time.Duration
	ServerAddress string
	BatchSize     int
	MaxRetries    int
	ID            uint32
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	bets   io.ReadCloser
	done   chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(bets io.ReadCloser, config ClientConfig) *Client {
	client := &Client{
		config: config,
		bets:   bets,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
}

func (c *Client) HandleSignals(ch <-chan os.Signal, done chan struct{}) {
	c.done = done
	go func() {
		signal := <-ch
		log.Infof("action: signal | result: success | client_id: %v | msg: received %s", c.config.ID, signal)
		close(c.done)
	}()

}

type ServerResponse struct {
	Response protocol.Response
	Err      error
}

type HandleContext struct {
	Results       chan<- ServerResponse
	Done          <-chan struct{}
	Requests      <-chan protocol.Request
	Timeout       time.Duration
	Backoff       time.Duration
	MaxRetries    int
	ServerAddress string
	ID            uint32
}

var ErrUnexpectedResponse = errors.New("unexpected response")

func HandleBets(ctx *HandleContext, conn net.Conn) {
	defer conn.Close()
	for req := range ctx.Requests {
		conn.SetWriteDeadline(time.Now().Add(ctx.Timeout))
		if err := protocol.EncodeRequest(conn, req); err != nil {
			ctx.Results <- ServerResponse{Err: err}
			return
		}
		conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
		res, err := protocol.DecodeResponse(conn)
		if err != nil {
			ctx.Results <- ServerResponse{Err: err}
			return
		}
		ack, ok := res.(protocol.Acknowledge)
		if !ok {
			ctx.Results <- ServerResponse{Err: errors.Wrap(ErrUnexpectedResponse, "expected acknowledge")}
		} else {
			ctx.Results <- ServerResponse{Response: ack}
		}
	}

	req := protocol.Done{ID: ctx.ID}
	conn.SetWriteDeadline(time.Now().Add(ctx.Timeout))
	if err := protocol.EncodeRequest(conn, req); err != nil {
		ctx.Results <- ServerResponse{Err: err}
		return
	}

	GetWinners(ctx)

	return
}

const BackoffExp = 2

func tryGetWinners(conn net.Conn, ID uint32, timeout time.Duration) (protocol.Response, error) {
	defer conn.Close()
	req := protocol.Winners{ID: ID}
	conn.SetWriteDeadline(time.Now().Add(timeout))
	if err := protocol.EncodeRequest(conn, req); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	res, err := protocol.DecodeResponse(conn)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetWinners(ctx *HandleContext) {
	retries := 0
	backoff := ctx.Backoff

	for retries < ctx.MaxRetries {
		conn, err := net.Dial("tcp", ctx.ServerAddress)
		if err != nil {
			ctx.Results <- ServerResponse{Err: err}
			return
		}
		select {
		case <-ctx.Done:
			conn.Close()
			return
		default:
			res, err := tryGetWinners(conn, ctx.ID, ctx.Timeout)
			if err != nil {
				ctx.Results <- ServerResponse{Err: err}
				return
			}
			switch r := res.(type) {
			case protocol.WinnersUnavailable:
				time.Sleep(backoff)
				retries++
				backoff *= BackoffExp
				backoff += time.Duration(rand.Int63n(100)) * time.Millisecond
			case protocol.WinnersList:
				ctx.Results <- ServerResponse{Response: r}
				return
			default:
				err := errors.Wrap(ErrUnexpectedResponse, "expected winners_list or winners_unavailable messages")
				ctx.Results <- ServerResponse{Err: err}
				return
			}
		}
	}
	err := errors.Errorf("couldn't get winners: max retry attemps reached %d", ctx.MaxRetries)
	ctx.Results <- ServerResponse{Err: err}
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	c.createClientSocket()
	defer func() {
		c.bets.Close()
	}()

	resultChannel := make(chan ServerResponse)
	requestChannel := make(chan protocol.Request)

	batchCtx := &BatchContext{
		ID:        c.config.ID,
		Reader:    c.bets,
		BatchSize: c.config.BatchSize,
		Requests:  requestChannel,
	}
	go HandleBatchs(batchCtx)

	handleCtx := &HandleContext{
		Results:       resultChannel,
		Done:          c.done,
		Requests:      requestChannel,
		Timeout:       c.config.Timeout,
		Backoff:       c.config.Backoff,
		MaxRetries:    c.config.MaxRetries,
		ServerAddress: c.config.ServerAddress,
		ID:            c.config.ID,
	}

	go HandleBets(handleCtx, c.conn)

	for {
		select {
		case res := <-resultChannel:
			if res.Err != nil {
				log.Errorf("action: error_detected | result: success | client_id: %v | error: %s",
					c.config.ID, res.Err.Error())
				if !errors.Is(res.Err, ErrUnexpectedResponse) {
					return
				}
			}
			switch r := res.Response.(type) {
			case protocol.Acknowledge:
				log.Infof("action: apuestas_enviadas | result: success | cantidad: %d", r.BetCount)
			case protocol.WinnersUnavailable:
				log.Info("action: consulta_ganadores | result: fail")
			case protocol.WinnersList:
				log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", r.WinnerCount)
				if r.WinnerCount != 0 {
					log.Infof("action: consulta_ganadores | result: success | docs_ganadores: %v", r.DNIS)
				}
				return
			}
		case <-c.done:
			return
		}
	}
}
