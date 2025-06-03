package proofs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	canine "github.com/jackalLabs/canine-chain/v4/app"

	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
	merkletree "github.com/wealdtech/go-merkletree/v2"
	"github.com/wealdtech/go-merkletree/v2/sha3"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ErrNotOurs  = "not our deal"
	ErrNotReady = "not ready yet"
)

func GenerateMerkleProof(tree *merkletree.MerkleTree, index int, item []byte) (bool, *merkletree.Proof, error) {
	log.Debug().Msg(fmt.Sprintf("Generating Merkle proof for index: %d", index))

	h := sha256.New()
	_, err := fmt.Fprintf(h, "%d%x", index, item)
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

// GenProof generates a proof from an arbitrary file on the file system
//
// returns proof, item and error
func GenProof(io FileSystem, merkle []byte, owner string, start int64, block int, chunkSize int, proofType int64) ([]byte, []byte, error) {
	log.Debug().Msg(fmt.Sprintf("About to generate merkle proof for file: %x", merkle))

	tree, chunk, err := io.GetFileTreeByChunk(merkle, owner, start, block, chunkSize, proofType)
	if err != nil {
		e := fmt.Errorf("cannot get chunk for %x at %d | %w", merkle, block, err)
		log.Error().Err(e)
		return nil, nil, e
	}

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
		return nil, nil, fmt.Errorf("failed to marshal proof %w", err)
	}

	log.Debug().Msg(fmt.Sprintf("Done making proof for %x", merkle))

	return jproof, chunk, nil
}

func (p *Prover) GenerateProof(merkle []byte, owner string, start int64, blockHeight int64, startedAt time.Time) ([]byte, []byte, int64, error) {
	log.Debug().Msg(fmt.Sprintf("Generating proof for %x", merkle))
	queryParams := &types.QueryFile{
		Merkle: merkle,
		Owner:  owner,
		Start:  start,
	}

	res, err := p.query.File(context.Background(), queryParams)
	if err != nil {
		return nil, nil, 0, err
	}

	file := res.File

	proofQuery := &types.QueryProof{
		ProviderAddress: p.wallet.AccAddress(),
		Merkle:          file.Merkle,
		Owner:           file.Owner,
		Start:           file.Start,
	}

	newProof := types.FileProof{ // defining a new proof model
		Prover:       p.wallet.AccAddress(),
		Merkle:       file.Merkle,
		Owner:        file.Owner,
		Start:        file.Start,
		LastProven:   0,
		ChunkToProve: 0,
	}

	if file.ContainsProver(p.wallet.AccAddress()) {
		// file is ours
		proofRes, err := p.query.Proof(context.Background(), proofQuery)
		if err == nil {
			newProof = proofRes.Proof // found the proof, we're good to go
		}
	} else {
		// file is not ours, we need to figure out what to do with it
		if len(file.Proofs) == int(file.MaxProofs) {
			return nil, nil, 0, errors.New(ErrNotOurs) // there is no more room on this file anyway, ignore it
		}
	}

	log.Debug().Msg(fmt.Sprintf("Querying proof of %x", merkle))

	t := time.Since(startedAt)

	proven := file.ProvenThisBlock(blockHeight+int64(t.Seconds()/5.0), newProof.LastProven)
	if proven {
		log.Debug().Msg(fmt.Sprintf("%x was proven at %d, height is now %d", file.Merkle, newProof.LastProven, blockHeight))
		return nil, nil, 0, nil
	}
	log.Info().Msg(fmt.Sprintf("%x was not yet proven at %d, height is now %d", file.Merkle, newProof.LastProven, blockHeight))

	block := int(newProof.ChunkToProve)

	log.Debug().Msg(fmt.Sprintf("Getting file tree by chunk for %x", merkle))

	proof, item, err := GenProof(p.io, merkle, owner, start, block, p.chunkSize, file.ProofType)

	return proof, item, newProof.ChunkToProve, err
}

func (p *Prover) PostProof(merkle []byte, owner string, start int64, blockHeight int64, startedAt time.Time) error {
	proof, item, index, err := p.GenerateProof(merkle, owner, start, blockHeight, startedAt)
	p.Dec()
	filesProving.Dec()
	if err != nil {
		log.Error().Msgf("Failed to generate proof for %x at %d", merkle, index)
		log.Error().Msg(err.Error())
		return err
	}

	if proof == nil || item == nil {
		log.Debug().Msg("generated proof was nil but no error was thrown")
		return nil
	}

	log.Debug().Msg("Successfully generated proof")

	msg := types.NewMsgPostProof(p.wallet.AccAddress(), merkle, owner, start, item, proof, index)

	m, wg := p.q.Add(msg)

	wg.Wait()

	if m.Error() != nil {
		log.Error().Err(m.Error())
		return m.Error()
	}

	var postRes types.MsgPostProofResponse
	data, err := hex.DecodeString(m.Res().Data)
	if err != nil {
		return nil
	}

	encodingCfg := canine.MakeEncodingConfig()
	var txMsgData sdk.TxMsgData
	err = encodingCfg.Marshaler.Unmarshal(data, &txMsgData)
	if err != nil {
		return nil
	}

	if len(txMsgData.Data) == 0 {
		return nil
	}

	err = postRes.Unmarshal(txMsgData.Data[m.Index()].Data)
	if err != nil {
		return nil
	}

	if !postRes.Success {
		log.Error().Msg(postRes.ErrorMessage)
	}

	log.Debug().Msg(fmt.Sprintf("%x was successfully proven", merkle))

	log.Debug().Msg(fmt.Sprintf("TX Hash: %s", m.Hash()))

	return nil
}

func (p *Prover) Start() {
	p.running = true
	for p.running {
		if !p.running {
			return
		}

		time.Sleep(time.Millisecond * 1000) // pauses for one second
		if !p.processed.Add(time.Second * time.Duration(p.interval)).Before(time.Now()) {
			continue
		}

		log.Debug().Time("start at", p.processed).Msg("Starting proof cycle...")

		abciInfo, err := p.wallet.Client.RPCClient.ABCIInfo(context.Background())
		if err != nil {
			log.Error().Err(err)
			continue
		}
		height := abciInfo.Response.LastBlockHeight + 10

		t := time.Now()

		err = p.io.ProcessFiles(func(merkle []byte, owner string, start int64) {
			for p.Full() {
				log.Debug().Msg("Proving queue is full, waiting...")

				time.Sleep(time.Second * 5)
			}
			log.Debug().Hex("merkle", merkle).Str("owner", owner).Int64("start", start).Msg("proving file")
			filesProving.Inc()
			p.Inc()
			go p.wrapPostProof(merkle, owner, start, height, t)
		})
		if err != nil {
			log.Error().Err(err).Msg("something went wrong while processing files in this proving cycle")
		}

		log.Debug().Time("finish at", time.Now()).TimeDiff("duration", time.Now(), p.processed).Msg("End of proof cycle")
		p.processed = time.Now()
	}
	log.Info().Msg("Prover module stopped")
}

func (p *Prover) wrapPostProof(merkle []byte, owner string, start int64, height int64, startedAt time.Time) {
	err := p.PostProof(merkle, owner, start, height, startedAt)
	if err != nil {
		log.Warn().
			Err(err).
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Int64("height", height).
			Msg("proof error")

		if code := status.Code(err); code > codes.NotFound {
			log.Error().
				Hex("merkle", merkle).
				Str("owner", owner).
				Int64("start", start).
				Int64("height", height).
				Time("startedAt", startedAt).
				Err(err).
				Msg("problem with rpc node")
			return
		}
		if code := status.Code(err); code == codes.NotFound {
			log.Debug().
				Hex("merkle", merkle).
				Str("owner", owner).
				Int64("start", start).
				Msg("deleting the file that no longer exists on the network")

			err := p.io.DeleteFile(merkle, owner, start)
			if err != nil {
				log.Error().
					Err(err).
					Hex("merkle", merkle).
					Msg("failed to delete file that no longer exist on the network")
			}
		}
		if err.Error() == ErrNotOurs { // if the file is not ours, delete it
			log.Debug().
				Hex("merkle", merkle).
				Str("owner", owner).
				Int64("start", start).
				Msg("deleting the file that does not belong to this provider")

			err := p.io.DeleteFile(merkle, owner, start)
			if err != nil {
				log.Error().
					Hex("merkle", merkle).
					Err(err).
					Msg("failed to delete file that does not belong to this provider")
			}
		}
	}
}

func (p *Prover) Stop() {
	p.running = false
}

func NewProver(wallet *wallet.Wallet, query types.QueryClient, q queue.Queue, io FileSystem, interval int64, threads int16, chunkSize int) *Prover {
	p := Prover{
		running:        false,
		wallet:         wallet,
		query:          query,
		q:              q,
		processed:      time.Time{},
		interval:       interval,
		io:             io,
		threads:        int32(threads),
		currentThreads: atomic.Int32{},
		chunkSize:      chunkSize,
	}

	return &p
}
