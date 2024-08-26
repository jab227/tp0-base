package common

import (
	"os"
	"os/signal"
	"syscall"
)

type SignalHandler struct {
	signalCh <-chan os.Signal
	doneCh   chan struct{}
}

func (s SignalHandler) Signal() <-chan os.Signal {
	return s.signalCh
}

func (s SignalHandler) Done() <-chan struct{} {
	return s.doneCh
}

func (s SignalHandler) Run() {
	signal := <-s.signalCh
	log.Infof("action: signal | result: success | signal: received %s", signal)
	s.doneCh <- struct{}{}
}

func NewSignalHandler() SignalHandler {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	doneChannel := make(chan struct{}, 1)
	return SignalHandler{
		signalCh: signalChannel,
		doneCh:   doneChannel,
	}
}
