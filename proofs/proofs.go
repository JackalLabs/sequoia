package proofs

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/sha3"
	"io"
	"strconv"
	"time"
)

const ErrNotOurs = "not our deal"
const ErrNotReady = "not ready yet"

func GenerateMerkleProof(tree *merkletree.MerkleTree, index int, item []byte) (valid bool, proof *merkletree.Proof, err error) {
	h := sha256.New()
	_, err = io.WriteString(h, fmt.Sprintf("%d%x", index, item))
	if err != nil {
		return
	}

	proof, err = tree.GenerateProof(h.Sum(nil), 0)
	if err != nil {
		return
	}

	valid, err = merkletree.VerifyProofUsing(h.Sum(nil), false, proof, [][]byte{tree.Root()}, sha3.New512())
	return
}

func (p *Prover) GenerateProof(cid string) ([]byte, []byte, error) {
	queryParams := &types.QueryActiveDealRequest{
		Cid: cid,
	}

	cl := types.NewQueryClient(p.wallet.Client.GRPCConn)

	res, err := cl.ActiveDeals(context.Background(), queryParams)
	if err != nil { // if the deal doesn't exist we check strays & contracts, then remove it
		contractParams := &types.QueryContractRequest{
			Cid: cid,
		}
		_, e := cl.Contracts(context.Background(), contractParams)
		if e != nil { // if we can't find the contract, check strays, then remove it
			strayParams := &types.QueryStrayRequest{
				Cid: cid,
			}
			_, e := cl.Strays(context.Background(), strayParams)
			if e != nil { // if we can't find the stray, remove it
				return nil, nil, err
			}

			return nil, nil, fmt.Errorf(ErrNotReady)
		}

		return nil, nil, fmt.Errorf(ErrNotReady)
	}

	if res.ActiveDeals.Provider != p.wallet.AccAddress() {
		return nil, nil, fmt.Errorf(ErrNotOurs)
	}

	if res.ActiveDeals.Proofverified == "true" {
		return nil, nil, nil
	}

	blockToProveString := res.ActiveDeals.Blocktoprove
	blockToProve, err := strconv.ParseInt(blockToProveString, 10, 64)
	if err != nil {
		return nil, nil, err
	}
	block := int(blockToProve)

	tree, chunk, err := file_system.GetFileChunk(p.db, cid, block)
	if err != nil {
		return nil, nil, err
	}

	valid, proof, err := GenerateMerkleProof(tree, block, chunk)
	if err != nil {
		return nil, nil, err
	}
	if !valid {
		return nil, nil, fmt.Errorf("tree not valid")
	}

	jproof, err := json.Marshal(*proof)
	if err != nil {
		return nil, nil, err
	}

	return jproof, chunk, nil
}

func (p *Prover) PostProof(cid string) error {
	proof, item, err := p.GenerateProof(cid)
	if err != nil {
		return err
	}

	if proof == nil {
		return nil
	}

	msg := types.NewMsgPostproof(p.wallet.AccAddress(), fmt.Sprintf("%x", item), string(proof), cid)

	_, _ = p.q.Add(msg)

	return nil

}

func (p *Prover) Start() {
	p.running = true
	for {
		if !p.running { // stop when running is false
			return
		}

		time.Sleep(time.Millisecond * 333)                                                // pauses for one third of a second
		if !p.processed.Add(time.Second * time.Duration(p.interval)).Before(time.Now()) { // check every 2 minutes
			continue
		}

		if p.locked {
			continue
		}

		p.locked = true

		files, err := file_system.ListFiles(p.db)
		if err != nil {
			fmt.Println(err)
		}

		for _, cid := range files {
			err := p.PostProof(cid)
			if err != nil {
				if err.Error() == "rpc error: code = NotFound desc = not found" { // if the file is not found on the network, delete it
					err := file_system.DeleteFile(p.db, cid)
					if err != nil {
						fmt.Println(err)
					}
				}
				if err.Error() == ErrNotOurs { // if the file is not ours, delete it
					err := file_system.DeleteFile(p.db, cid)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		p.locked = false

		p.processed = time.Now()
	}
}

func (p *Prover) Stop() {
	p.running = false
}

func NewProver(wallet *wallet.Wallet, db *badger.DB, q *queue.Queue, interval int64) *Prover {
	p := Prover{
		running:   false,
		wallet:    wallet,
		db:        db,
		q:         q,
		processed: time.Time{},
		interval:  interval,
	}

	return &p
}
