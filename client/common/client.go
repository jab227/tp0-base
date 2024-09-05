package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
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
	err error
}

type Request struct {
	m    protocol.Marshaler
	ID   uint32
	kind protocol.RequestKind
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
		// Receive batches, create request and send them to server
		log.Debug("starting to send requests")
		for incoming := range requests {
			req := protocol.NewBetRequest(incoming.kind, incoming.ID, incoming.m)
			id = incoming.ID
			// Send Request
			rw.SetWriteDeadline(time.Now().Add(timeout))			
			if err := protocol.EncodeRequest(w, req); err != nil {
				responseChannel <- ServerResponse{err: err}
				return
			}
			// Receive Response
			w.Flush()
			// Expect ack
			rw.SetReadDeadline(time.Now().Add(timeout))			
			_, err := expectAcknowledge(r)
			if err != nil {
				responseChannel <- ServerResponse{err: err}
				continue
			}

		}
		// Send end batches req
		req := protocol.NewBetRequest(protocol.BetBatchStop, id, nil)
		rw.SetWriteDeadline(time.Now().Add(timeout))
		if err := protocol.EncodeRequest(w, req); err != nil {
			responseChannel <- ServerResponse{err: err}
			return
		}
		w.Flush()
		// Process Response: expect ack
		rw.SetReadDeadline(time.Now().Add(timeout))
		_, err := expectAcknowledge(r)
		if err != nil {
			responseChannel <- ServerResponse{err: err}
			return
		}
		// Wait for ready
		backoff := utils.Backoff{
			Time:    time.Duration(500) * time.Millisecond,
			Retries: 10,
			Exp:     2,
		}
		readyPtr := new(protocol.Ready)
		task := func() (bool, error) {
			rw.SetReadDeadline(time.Now().Add(timeout))
			ready, err := expectReady(r)
			if err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					err = errors.Wrap(err, "couldn't receive ready")
					return true, err
				}
				err = errors.Wrap(err, "retrying")
				return false, err
			}
			*readyPtr = ready
			return true, nil
		}

		onError := func(err error) {
			log.Errorf("action: wait_ready | result: failure | client_id: %v | error: %", id, err)
		}
		if !backoff.Try(task, onError) {
			log.Errorf("action: wait_ready | result: failure | client_id: %v | error: ready not received", id, err)
			return
		}

		// Ask for winners
		winners_req := protocol.NewBetRequest(protocol.BetGetWinners, id, nil)
		rw.SetWriteDeadline(time.Now().Add(timeout))
		if err := protocol.EncodeRequest(w, winners_req); err != nil {
			responseChannel <- ServerResponse{err: err}
			return
		}
		w.Flush()
		rw.SetReadDeadline(time.Now().Add(timeout))
		winners, err := expectWinners(r)
		if err != nil {
			responseChannel <- ServerResponse{err: err}
			return
		}
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(winners.DNIs))
	}()
	return responseChannel
}

func expectAcknowledge(r *bufio.Reader) (protocol.BetAcknowledge, error) {
	var res protocol.BetResponse
	err := res.DecodeResponse(r)
	if err != nil {
		return protocol.BetAcknowledge{}, err
	}
	u := res.GetType()
	ack, ok := u.(*protocol.BetAcknowledge)
	if !ok {
		err := fmt.Errorf("wrong response: expected acknowledge")
		return protocol.BetAcknowledge{}, err
	}
	if err := ack.UnmarshalPayload(res.Payload); err != nil {
		err := errors.Wrap(err, "couldn't unmarshal acknowledge")
		return protocol.BetAcknowledge{}, err
	}
	return *ack, nil
}

func expectReady(r *bufio.Reader) (protocol.Ready, error) {
	var res protocol.BetResponse
	err := res.DecodeResponse(r)
	if err != nil {
		return protocol.Ready{}, err
	}
	u := res.GetType()
	ready, ok := u.(*protocol.Ready)
	if !ok {
		err := fmt.Errorf("wrong response: expected ready")
		return protocol.Ready{}, err
	}
	if err := ready.UnmarshalPayload(res.Payload); err != nil {
		err := errors.Wrap(err, "couldn't unmarshal ready")
		return protocol.Ready{}, err
	}
	return *ready, nil
}

func expectWinners(r *bufio.Reader) (protocol.Winners, error) {
	var res protocol.BetResponse
	err := res.DecodeResponse(r)
	if err != nil {
		return protocol.Winners{}, err
	}
	u := res.GetType()
	winners, ok := u.(*protocol.Winners)
	if !ok {
		err := fmt.Errorf("wrong response: expected winners")
		return protocol.Winners{}, err
	}
	if err := winners.UnmarshalPayload(res.Payload); err != nil {
		err := errors.Wrap(err, "couldn't unmarshal winners")
		return protocol.Winners{}, err
	}
	return *winners, nil
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
