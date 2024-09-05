package utils

import (
	"math/rand"
	"time"
)

type Backoff struct {
	Time    time.Duration
	Retries int
	Exp     int
}

type Task func() (bool, error)

func (b *Backoff) Try(t Task, onError func(err error)) bool {
	retries := 0
	backoff := b.Time
	for retries < b.Retries {
		finished, err := t()
		if err != nil {
			onError(err)
		}
		if finished {
			return true
		}
		time.Sleep(backoff)
		retries++
		backoff *= time.Duration(2)
		backoff += time.Duration(rand.Int63n(100)) * time.Millisecond
	}
	return false
}
