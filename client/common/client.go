package common

import (
	"io"
	"net"
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

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	c.createClientSocket()
	defer func() {
		c.conn.Close()
		c.bets.Close()
	}()

	resultChannel := make(chan HandlerResponse)
	requestChannel := make(chan protocol.Request)

	batchCtx := &BatchHandlerContext{
		ID:        c.config.ID,
		Reader:    c.bets,
		BatchSize: c.config.BatchSize,
		Requests:  requestChannel,
	}
	go HandleBetsBatching(batchCtx)

	handleCtx := &ProtocolHandlerContext{
		Results:       resultChannel,
		Done:          c.done,
		Requests:      requestChannel,
		Timeout:       c.config.Timeout,
		Backoff:       c.config.Backoff,
		MaxRetries:    c.config.MaxRetries,
		ServerAddress: c.config.ServerAddress,
		ID:            c.config.ID,
	}

	go HandleProtocol(handleCtx, c.conn)

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
				return
			}
		case <-c.done:
			return
		}
	}
}
