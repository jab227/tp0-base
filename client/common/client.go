package common

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	Timeout       time.Duration
	ServerAddress string
	BatchSize     int
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
		c.conn.Close()
		log.Infof("action: close_socket | result: success | client_id: %v | msg: closed client socket",
			c.config.ID)
		c.done <- struct{}{}
	}()

}

type ServerResponse struct {
	Response protocol.Response
	Err      error
}

type HandleContext struct {
	Conn     net.Conn
	Results  chan<- ServerResponse
	Done     chan<- struct{}
	Requests <-chan protocol.Request
	Timeout  time.Duration
}

func HandleBets(ctx *HandleContext) {
	for req := range ctx.Requests {
		ctx.Conn.SetWriteDeadline(time.Now().Add(ctx.Timeout))
		if err := protocol.EncodeRequest(ctx.Conn, req); err != nil {
			ctx.Results <- ServerResponse{Err: err}
			return
		}
		ctx.Conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
		ack, err := protocol.DecodeResponse(ctx.Conn)
		if err != nil {
			ctx.Results <- ServerResponse{Err: err}
			return
		}
		ctx.Results <- ServerResponse{Response: ack}
	}

	// This may fail but all the bets where sent
	ctx.Conn.SetWriteDeadline(time.Now().Add(ctx.Timeout))
	req := protocol.Done{}
	if err := protocol.EncodeRequest(ctx.Conn, req); err != nil {
		ctx.Results <- ServerResponse{Err: err}
		return
	}

	ctx.Done <- struct{}{}

	return
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	c.createClientSocket()
	defer func() {
		c.conn.Close()
		log.Infof("action: close_socket | result: success | client_id: %v | msg: closed client socket",
			c.config.ID)

		c.bets.Close()
		log.Infof("action: close_file | result: success | client_id: %v | msg: closed dataset file", c.config.ID)
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
		Conn:     c.conn,
		Results:  resultChannel,
		Done:     c.done,
		Requests: requestChannel,
		Timeout:  c.config.Timeout,
	}

	go HandleBets(handleCtx)

	for {
		select {
		case res := <-resultChannel:
			if res.Err != nil {
				log.Errorf("action: error_detected | result: success | client_id: %v | error: %s",
					c.config.ID, res.Err.Error(),
				)
				return
			}

			// CHECK(juan): Do I log all the bets?
			log.Infof("action: apuestas_enviadas | result: success | cantidad: %d", res.Response.BetCount)
		case <-c.done:
			return
		}
	}
}
