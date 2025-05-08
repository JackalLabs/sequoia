package queue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
)

func TestPoolSendToAny(t *testing.T) {
	r := require.New(t)
	queue := make([]chan *Message, 0)

	for range 4 {
		queue = append(queue, make(chan *Message))
	}

	pool := Pool{
		workerChannels: queue,
	}

	msg := Message{
		msg: types.NewMsgPostProof("", []byte("hello"), "owner", 0, []byte("item"), []byte("list"), 0),
		wg:  nil,
	}

	wg := sync.WaitGroup{}
	go func() {
		t.Logf("starting goroutine(worker): %d", len(queue)-1)
		for m := range queue[len(queue)-1] {
			t.Logf("worker[%d]: message received", len(queue)-1)
			_ = m
			wg.Done()
		}
	}()

	msgCount := 8
	for range msgCount {
		wg.Add(1)
		to := pool.sendToAny(&msg)
		r.EqualValues(len(queue)-1, to)
	}

	wg.Wait()
	close(queue[len(queue)-1])
}
