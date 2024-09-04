package common

import (
	"bufio"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
	"github.com/op/go-logging"
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
	SocketTimeout time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	chunks <-chan BatchResult
	doneCh <-chan struct{}
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, chunks <-chan BatchResult, done <-chan struct{}) *Client {
	client := &Client{
		doneCh: done,
		config: config,
		chunks: chunks,
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

type Request struct {
	m    protocol.Marshaler
	ID   uint32
	kind protocol.MessageKind
}

// TODO(JUAN) change to SendRequest
func SendRequests(rw utils.DeadlineReadWriter, requests <-chan Request, timeout time.Duration) <-chan ServerResponse {
	//
	w := bufio.NewWriter(rw)
	r := bufio.NewReader(rw)
	responseChannel := make(chan ServerResponse, 1)
	go func() {
		var id uint32
		defer close(responseChannel)
		for incoming := range requests {
			req := protocol.NewBetRequest(incoming.kind, incoming.ID, incoming.m)
			id = incoming.ID

			rw.SetWriteDeadline(time.Now().Add(timeout))
			if err := protocol.EncodeRequest(w, req); err != nil {
				responseChannel <- ServerResponse{err: err}
				return
			}

			w.Flush()
			rw.SetReadDeadline(time.Now().Add(timeout))
			ack, err := protocol.DecodeResponse(r)
			if err != nil {
				responseChannel <- ServerResponse{err: err}
				return
			}
			responseChannel <- ServerResponse{ack: ack}
		}
		req := protocol.NewBetRequest(protocol.BetBatchStop, id, nil)
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
func (c *Client) StartClientLoop(p *BatchProcessor) {
	if err := c.createClientSocket(); err != nil {
		log.Fatalf("couldn't create client socket: %s", err)
	}
	log.Debug("client socket created")
	defer func() {
		if err := c.conn.Close(); err != nil {
			log.Errorf("action: close_socket | result: failure | client_id: %v | error: %s", c.config.ID, err)
		} else {
			log.Infof("action: close_socket | result: success | client_id: %v | msg: closed client socket",
				c.config.ID)
		}
	}()

	id, err := strconv.Atoi(c.config.ID)
	if err != nil {
		log.Fatalf("couldn't parse agency id from env data: %s", err)
	}
	requests := make(chan Request, 1)
	go func() {
		id := uint32(id)
		defer close(requests)
		for r := range c.chunks {
			if r.Err != nil {
				log.Errorf("couldn't create request: %s", err)
				continue
			}
			requests <- Request{ID: id, kind: protocol.BetBatch, m: r.Chunk}
		}
	}()

	responseCh := SendRequests(c.conn, requests, c.config.SocketTimeout)
	for {
		select {
		case res, more := <-responseCh:
			if !more {
				return
			}
			if res.err != nil {
				log.Errorf("action: apuesta_enviada | result: failure | client_id: %v | error: %s",
					c.config.ID, res.err,
				)
			}
		case <-c.doneCh:
			p.Stop()
			return
		}
	}
}
