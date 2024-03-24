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
	LoopLapse     time.Duration
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
	ack protocol.Ack
	err error
}

func SendBet(timeout time.Duration, conn net.Conn, result chan<- ServerResponse, requests <-chan protocol.Request) {
	for req := range requests {
		conn.SetReadDeadline(time.Now().Add(timeout))
		if err := protocol.EncodeRequest(conn, req); err != nil {
			result <- ServerResponse{err: err}
			return
		}

		conn.SetWriteDeadline(time.Now().Add(timeout))
		ack, err := protocol.DecodeResponse(conn)
		if err != nil {
			result <- ServerResponse{err: err}
			return
		}
		result <- ServerResponse{ack: ack}
	}
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	c.createClientSocket()
	defer c.conn.Close()

	resultChannel := make(chan ServerResponse)
	requestChannel := make(chan protocol.Request)

	go BatchReader(c.config.ID, c.bets, c.config.BatchSize, requestChannel)
	go SendBet(c.config.LoopLapse, c.conn, resultChannel, requestChannel)

	for {
		select {
		case res := <-resultChannel:
			if res.err != nil {
				log.Errorf("action: error_detected | result: success | client_id: %v | error: %s",
					c.config.ID, res.err.Error(),
				)
				break
			}
			// CHECK(juan): Do I log all the bets?
			log.Infof("action: apuestas_enviadas | result: success | cantidad: %d", res.ack.BetCount)
		case <-c.done:
			return
		}
	}
}
