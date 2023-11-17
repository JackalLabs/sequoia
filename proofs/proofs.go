package proofs

import (
	"context"
	"crypto/sha256"
	"errors"
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

func GenerateMerkleProof(tree *merkletree.MerkleTree, index int, item []byte) (bool, *merkletree.Proof, error) {
	log.Debug().Msg(fmt.Sprintf("Generating Merkle proof for %d", index))

	h := sha256.New()
	_, err := io.WriteString(h, fmt.Sprintf("%d%x", index, item))
	if err != nil {
		return false, nil, err
	}

	proof, err := tree.GenerateProof(h.Sum(nil), 0)
	if err != nil {
		return false, nil, err
	}

	valid, err := merkletree.VerifyProofUsing(h.Sum(nil), false, proof, [][]byte{tree.Root()}, sha3.New512())
	if err != nil {
		return false, nil, err
	}
	return valid, proof, nil
}

func (p *Prover) GenerateProof(merkle []byte, owner string, start int64, blockHeight int64, startedAt time.Time) ([]byte, []byte, error) {
	log.Debug().Msg(fmt.Sprintf("Generating proof for %x", merkle))
	queryParams := &types.QueryFile{
		Merkle: merkle,
		Owner:  owner,
		Start:  start,
	}

	cl := types.NewQueryClient(p.wallet.Client.GRPCConn)

	res, err := cl.File(context.Background(), queryParams)
	if err != nil {
		return nil, nil, err
	}

	file := res.File

	proofQuery := &types.QueryProof{
		ProviderAddress: p.wallet.AccAddress(),
		Merkle:          file.Merkle,
		Owner:           file.Owner,
		Start:           file.Start,
	}

	var newProof types.FileProof
	log.Debug().Msg(fmt.Sprintf("Querying proof of %x", merkle))

	proofRes, err := cl.Proof(context.Background(), proofQuery)
	if err != nil {
		if len(file.Proofs) == int(file.MaxProofs) {
			return nil, nil, errors.New(ErrNotOurs)
		}
		newProof = types.FileProof{
			Prover:       p.wallet.AccAddress(),
			Merkle:       file.Merkle,
			Owner:        file.Owner,
			Start:        file.Start,
			LastProven:   0,
			ChunkToProve: 0,
		}
	} else { // if the proof does exist we handle it
		newProof = proofRes.Proof
	}

	t := time.Since(startedAt)

	proven := file.ProvenThisBlock(blockHeight+int64(t.Seconds()/6), newProof.LastProven)
	if proven {
		log.Info().Msg(fmt.Sprintf("%x was proven at %d, height is now %d", file.Merkle, newProof.LastProven, blockHeight))
		log.Debug().Msg("File was already proven")
		return nil, nil, nil
	}
	log.Info().Msg(fmt.Sprintf("%x was not yet proven at %d, height is now %d", file.Merkle, newProof.LastProven, blockHeight))

	block := int(newProof.ChunkToProve)

	log.Debug().Msg(fmt.Sprintf("Getting file tree by chunk for %x", merkle))

	tree, chunk, err := file_system.GetFileTreeByChunk(p.db, merkle, owner, start, block)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to get filetree by chunk for %x %w", merkle, err))
		return nil, nil, err
	}

	log.Debug().Msg(fmt.Sprintf("About to generate merkle proof for %x", merkle))

	valid, proof, err := GenerateMerkleProof(tree, block, chunk)
	if err != nil {
		return nil, nil, err
	}
	if !valid {
		e := errors.New("tree not valid")
		log.Error().Err(fmt.Errorf("tree not valid for %x %w", merkle, e))
		return nil, nil, e
	}

	jproof, err := json.Marshal(*proof)
	if err != nil {
		return nil, nil, err
	}

	log.Debug().Msg(fmt.Sprintf("Done making proof for %x", merkle))

	return jproof, chunk, nil
}

func (p *Prover) PostProof(merkle []byte, owner string, start int64, blockHeight int64, startedAt time.Time) error {
	proof, item, err := p.GenerateProof(merkle, owner, start, blockHeight, startedAt)
	if err != nil {
		return err
	}

	if proof == nil || item == nil {
		log.Debug().Msg("generated proof was nil but no error was thrown")
		return nil
	}

	log.Debug().Msg("Successfully generated proof")

	msg := types.NewMsgPostProof(p.wallet.AccAddress(), merkle, owner, start, item, proof)

	m, wg := p.q.Add(msg)

	wg.Wait()

	if m.Error() != nil {
		log.Error().Err(m.Error())
		return m.Error()
	}

	log.Debug().Msg(fmt.Sprintf("TX Hash: %s", m.Hash()))

	return nil
}

func (p *Prover) Start() {
	p.running = true
	for {
		if !p.running { // stop when running is false
			return
		}

		time.Sleep(time.Millisecond * 1000) // pauses for one third of a second
		if !p.processed.Add(time.Second * time.Duration(p.interval)).Before(time.Now()) {
			continue
		}

		log.Debug().Msg("Starting proof cycle...")

		abciInfo, err := p.wallet.Client.RPCClient.ABCIInfo(context.Background())
		if err != nil {
			log.Error().Err(err)
			continue
		}
		height := abciInfo.Response.LastBlockHeight + 1

		t := time.Now()

		err = file_system.ProcessFiles(p.db, func(merkle []byte, owner string, start int64) {
			log.Debug().Msg(fmt.Sprintf("proving: %x", merkle))
			go p.wrapPostProof(merkle, owner, start, height, t)
		})
		if err != nil {
			log.Error().Err(err)
		}

		p.processed = time.Now()
	}
}

func (p *Prover) wrapPostProof(merkle []byte, owner string, start int64, height int64, startedAt time.Time) {
	filesProving.Inc()
	defer filesProving.Dec()
	err := p.PostProof(merkle, owner, start, height, startedAt)
	if err != nil {
		log.Warn().Err(err)
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
