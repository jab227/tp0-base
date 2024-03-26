package protocol_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

func TestMarshalBatch(t *testing.T) {
	t.Run("marshal batch", func(t *testing.T) {
		const (
			numberOfBets = 5
			batchSize    = 1
		)
		bets := `Santiago Lionel,Lorca,30904465,1999-03-17,2201
Agustin Emanuel,Zambrano,21689196,2000-05-10,9325
Tiago Nicolás,Rivera,34407251,2001-08-29,1033
Camila Rocio,Varela,37130775,1995-05-09,4179
Diego Agustin,Mamani,33259835,1991-01-08,1931`
		r := strings.NewReader(bets)

		payloads := strings.Split(bets, "\n")
		batcher := common.NewBatcher(r, batchSize)
		finished, err := false, error(nil)
		i := 0
		for {
			finished, err = batcher.ReadBatch()
			if finished {
				break
			}
			if err != nil {
				t.Errorf("Unexpected error %s", err.Error())
			}
			payload, batched := batcher.MarshalBet()
			if batched != batchSize {
				t.Errorf("got %d, want %d", batched, batchSize)
			}
			expectedPayload := []byte(payloads[i])
			if !reflect.DeepEqual(payload, expectedPayload) {
				t.Errorf("got %q, want %q", string(payload), string(expectedPayload))
			}
			i++
		}

		if i != numberOfBets {
			t.Errorf("got %d, want %d", i, numberOfBets)
		}

	})
	t.Run("Smaller batch size if the total size is greater than 8KB", func(t *testing.T) {
		s := "Santiago Lionel,Lorca,30904465,1999-03-17,2201\n"
		nTimes := (common.MaxBatchByteSize / len(s)) + 2
		s = strings.Repeat(s, nTimes)
		s = s[:len(s) - 1]
		r := strings.NewReader(s)
		batcher := common.NewBatcher(r, nTimes)
		finished, err := false, error(nil)
		for {
			finished, err = batcher.ReadBatch()
			if finished {
				break
			}
			if err != nil {
				t.Errorf("Unexpected error %s", err.Error())
			}
			payload, batched := batcher.MarshalBet()
			fmt.Println(batched)
			if batched >= nTimes {
				t.Errorf("got %d, want %d", batched, nTimes)
			}
			if len(payload) > common.MaxBatchByteSize {
				t.Errorf("Got payload size %d, want %d", len(payload), common.MaxBatchByteSize)
			}
		}
	})
}
