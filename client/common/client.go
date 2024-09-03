package common

import (
	"bufio"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/op/go-logging"
	"io"
	"net"
	"strconv"
	"time"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	bettor agency.Bettor
	config ClientConfig
	chunks <-chan BatchResult
	doneCh <-chan struct{}
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, bettor agency.Bettor, chunks <-chan BatchResult, done <-chan struct{}) *Client {
	client := &Client{
		bettor: bettor,
		doneCh: done,
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

type ServerResponse struct {
	ack protocol.BetAcknowledge
	err error
}

// TODO(JUAN) change to SendRequest
func SendBet(ID uint32, bet agency.Bet, rw io.ReadWriter) <-chan ServerResponse {
	req := protocol.NewBetRequest(ID, bet)
	w := bufio.NewWriter(rw)
	r := bufio.NewReader(rw)
	responseChannel := make(chan ServerResponse, 1)
	go func() {
		if err := protocol.EncodeRequest(w, req); err != nil {
			responseChannel <- ServerResponse{err: err}
			return
		}
		w.Flush()
		ack, err := protocol.DecodeResponse(r)
		if err != nil {
			responseChannel <- ServerResponse{err: err}
			return
		}
		responseChannel <- ServerResponse{ack: ack}
	}()
	return responseChannel
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	if err := c.createClientSocket(); err != nil {
		log.Fatalf("couldn't create client socket: %s", err)
	}
	defer func() {
		if err := c.conn.Close(); err != nil {
			log.Errorf("action: close_socket | result: failure | client_id: %v | error: %s", c.config.ID, err)
		} else {
			log.Infof("action: close_socket | result: success | client_id: %v | msg: closed client socket",
				c.config.ID)
		}
	}()

	bet, err := agency.NewBet(c.bettor)
	if err != nil {
		log.Fatalf("couldn't create bettor from env data: %s", err)
	}
	id, err := strconv.Atoi(c.config.ID)
	if err != nil {
		log.Fatalf("couldn't parse agency id from env data: %s", err)
	}
	responseCh := SendBet(uint32(id), bet, c.conn)
	select {
	case res := <-responseCh:
		if res.err != nil {
			log.Errorf("action: apuesta_enviada | result: failure | client_id: %v | error: %s",
				c.config.ID, res.err,
			)
		} else {
			log.Infof("action: apuesta_enviada | result: success | documento: %s | numero: %d", c.bettor.DNI, res.ack.BetNumber)
		}
	case <-c.doneCh:
		return
	case <-time.After(c.config.LoopPeriod):
		log.Infof("action: timeout_detected | result: success | client_id: %v",
			c.config.ID,
		)
		return
	}
	log.Infof("action: client_finished | result: success | client_id: %v", c.config.ID)
}
