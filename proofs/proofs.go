package proofs

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"
	"github.com/wealdtech/go-merkletree/v2"
	"github.com/wealdtech/go-merkletree/v2/sha3"
)
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ErrNotOurs  = "not our deal"
	ErrNotReady = "not ready yet"
)

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

func (p *Prover) GenerateProof(merkle []byte, owner string, start int64, blockHeight int64) ([]byte, []byte, error) {
	queryParams := &types.QueryFile{
		Merkle: merkle,
		Owner:  owner,
		Start:  start,
	}

	cl := types.NewQueryClient(p.wallet.Client.GRPCConn)

	res, err := cl.File(context.Background(), queryParams)
	if err != nil { // if the deal doesn't exist we check strays & contracts, then remove it
		return nil, nil, err
	}

	file := res.File

	proofQuery := &types.QueryProof{
		ProviderAddress: p.wallet.AccAddress(),
		Merkle:          merkle,
		Owner:           owner,
		Start:           start,
	}

	var newProof types.FileProof

	proofRes, err := cl.Proof(context.Background(), proofQuery)
	if err != nil {
		if len(file.Proofs) == int(file.MaxProofs) {
			return nil, nil, fmt.Errorf(ErrNotOurs)
		}
		newProof.Owner = file.Owner
		newProof.Merkle = file.Merkle
		newProof.Start = file.Start
		newProof.Prover = p.wallet.AccAddress()
		newProof.ChunkToProve = 0
		newProof.LastProven = 0
	} else { // if the proof does exist we handle it
		newProof = proofRes.Proof
	}

	windowStart := blockHeight - (blockHeight % file.ProofInterval)

	if newProof.LastProven > windowStart { // already proven
		return nil, nil, nil
	}

	block := int(newProof.ChunkToProve)

	tree, chunk, err := file_system.GetFileTreeByChunk(p.db, merkle, owner, start, block)
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

func (p *Prover) PostProof(merkle []byte, owner string, start int64, blockHeight int64) error {
	proof, item, err := p.GenerateProof(merkle, owner, start, blockHeight)
	if err != nil {
		return err
	}

	if proof == nil {
		return nil
	}

	msg := types.NewMsgPostProof(p.wallet.AccAddress(), merkle, owner, start, item, proof)

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

		log.Debug().Msg("Starting proof cycle...")

		p.locked = true

		merkles, owners, starts, err := file_system.ListFiles(p.db)
		if err != nil {
			log.Error().Err(err)
		}

		abciInfo, err := p.wallet.Client.RPCClient.ABCIInfo(context.Background())
		if err != nil {
			log.Error().Err(err)
			continue
		}

		height := abciInfo.Response.LastBlockHeight

		for i, merkle := range merkles {
			owner := owners[i]
			start := starts[i]
			err := p.PostProof(merkle, owner, start, height)
			if err != nil {
				if err.Error() == "rpc error: code = NotFound desc = not found" { // if the file is not found on the network, delete it
					err := file_system.DeleteFile(p.db, merkle, owner, start)
					if err != nil {
						log.Error().Err(err)
					}
				}
				if err.Error() == ErrNotOurs { // if the file is not ours, delete it
					err := file_system.DeleteFile(p.db, merkle, owner, start)
					if err != nil {
						log.Error().Err(err)
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
