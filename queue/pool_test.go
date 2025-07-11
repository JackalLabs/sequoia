package queue

import (
	"testing"

	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
)

func TestPoolAddMsg(t *testing.T) {
	queue := make([]chan *Message, 0)

	for range 2 {
		queue = append(queue, make(chan *Message))
	}

	wallet, _, _, _ := setupWalletClient(t)

	workers := []*worker{{wallet: wallet}, {wallet: wallet}}

	pool := Pool{
		offsets:     workers,
		offsetQueue: queue,
	}

	go func() {
		msg := <-queue[0] // worker 0 is "busy" after receiving from queue[0]
		msg.wg.Done()
		msg = <-queue[1]
		msg.wg.Done()
	}()

	msg, wg0 := pool.Add(&types.MsgPostProof{})
	if _, ok := msg.msg.(*types.MsgPostProof); !ok {
		t.Errorf("expected: msg.msg as *types.MsgPostProof, got: %T", msg.msg)
	}
	wg0.Wait()

	msg, wg1 := pool.Add(&types.MsgPostProof{})
	if _, ok := msg.msg.(*types.MsgPostProof); !ok {
		t.Errorf("expected: msg.msg as *types.MsgPostProof, got: %T", msg.msg)
	}
	wg1.Wait()
}
