package common

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	ID            uint32
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	bettor agency.Bettor
	done   chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(bettor agency.Bettor, config ClientConfig) *Client {
	client := &Client{
		config: config,
		bettor: bettor,
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

// TODO(juan): Introduce buffering to both the encoder and the decoder
func SendBet(agencyID uint32, bet agency.Bet, rw io.ReadWriter, result chan<- ServerResponse) {
	req := protocol.NewBetRequest(agencyID, bet)
	if err := protocol.EncodeRequest(rw, req); err != nil {
		result <- ServerResponse{err: err}
		return
	}
	ack, err := protocol.DecodeResponse(rw)
	if err != nil {
		result <- ServerResponse{err: err}
		return
	}
	result <- ServerResponse{ack: ack}
	close(result)
	return
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	c.createClientSocket()
	defer c.conn.Close()
	resultChannel := make(chan ServerResponse)
	bet, err := agency.NewBet(c.bettor)
	if err != nil {
		log.Fatalf("couldn't parse bet from env: %s", err.Error())
	}
	go SendBet(c.config.ID, bet, c.conn, resultChannel)
	select {
	case res := <-resultChannel:
		if res.err != nil {
			log.Errorf("action: error_detected | result: success | client_id: %v | error: %s",
				c.config.ID, res.err.Error(),
			)
		} else {
			log.Infof("action: apuesta_enviada | result: success | documento: %s | numero: %d", c.bettor.DNI, res.ack.BetNumber)
		}
		return
	case <-c.done:
		return
	case <-time.After(c.config.LoopLapse):
		log.Infof("action: timeout_detected | result: success | client_id: %v",
			c.config.ID,
		)
		return
	}
}
