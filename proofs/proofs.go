package proofs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	sequoiaTypes "github.com/JackalLabs/sequoia/types"
	treeblake3 "github.com/wealdtech/go-merkletree/v2/blake3"
	"github.com/zeebo/blake3"

	"github.com/dgraph-io/badger/v4"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	canine "github.com/jackalLabs/canine-chain/v5/app"

	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v5/x/storage/types"
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

func GenerateMerkleProof(tree *merkletree.MerkleTree, index int, item []byte, proofType int64) (bool, *merkletree.Proof, error) {
	log.Debug().Msg(fmt.Sprintf("Generating Merkle proof for %d", index))

	h := sha256.New()
	if proofType == sequoiaTypes.ProofTypeBlake3 {
		h = blake3.New()
	}

	_, err := fmt.Fprintf(h, "%d%x", index, item)
	if err != nil {
		return false, nil, err
	}

	proof, err := tree.GenerateProof(h.Sum(nil), 0)
	if err != nil {
		return false, nil, err
	}

	var treeHash merkletree.HashType = sha3.New512()
	if proofType == sequoiaTypes.ProofTypeBlake3 {
		treeHash = treeblake3.New256()
	}

	valid, err := merkletree.VerifyProofUsing(h.Sum(nil), false, proof, [][]byte{tree.Root()}, treeHash)
	if err != nil {
		return false, nil, err
	}
	return valid, proof, nil
}

// GenProof generates a proof from an arbitrary file on the file system
//
// returns proof, item and error
func GenProof(io FileSystem, merkle []byte, owner string, start int64, block int, chunkSize int, proofType int64) ([]byte, []byte, error) {
	tree, chunk, err := io.GetFileTreeByChunk(merkle, owner, start, block, chunkSize, proofType)
	if err != nil {
		e := fmt.Errorf("cannot get chunk for %x at %d | %w", merkle, block, err)
		log.Error().Err(e)
		return nil, nil, e
	}

	log.Debug().Msg(fmt.Sprintf("About to generate merkle proof for %x", merkle))

	valid, proof, err := GenerateMerkleProof(tree, block, chunk, proofType)
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

	cl := types.NewQueryClient(p.wallet.Client.GRPCConn)

	res, err := cl.File(context.Background(), queryParams)
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
		proofRes, err := cl.Proof(context.Background(), proofQuery)
		if err == nil {
			newProof = proofRes.Proof // found the proof, we're good to go
		}
	} else {
		// file is not ours, we need to figure out what to do with it
		if len(file.Proofs) == int(file.MaxProofs) {
			// disable not ours check
			// return nil, nil, 0, errors.New(ErrNotOurs) // there is no more room on this file anyway, ignore it
			return nil, nil, 0, nil // there is no more room on this file anyway, ignore it
		}
	}

	log.Debug().Msg(fmt.Sprintf("Querying proof of %x", merkle))

	t := time.Since(startedAt)

	proven := file.ProvenThisBlock(blockHeight+int64(t.Seconds()/6.0), newProof.LastProven)
	if proven {
		log.Debug().Msg(fmt.Sprintf("%x was already proven at %d, height is now %d", file.Merkle, newProof.LastProven, blockHeight))
		return nil, nil, 0, nil
	}
	log.Debug().Msg(fmt.Sprintf("%x was not yet proven at %d, height is now %d", file.Merkle, newProof.LastProven, blockHeight))

	block := int(newProof.ChunkToProve)

	log.Debug().Msg(fmt.Sprintf("Getting file tree by chunk for %x", merkle))

	proof, item, err := GenProof(p.io, merkle, owner, start, block, p.chunkSize, file.ProofType)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("could not gen proof: %w", err)
	}

	return proof, item, newProof.ChunkToProve, err
}

func (p *Prover) PostProof(merkle []byte, owner string, start int64, blockHeight int64, startedAt time.Time) error {
	proof, item, index, err := p.GenerateProof(merkle, owner, start, blockHeight, startedAt)
	p.Dec()
	filesProving.Dec()
	if err != nil {
		log.Error().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Err(err).
			Msg("Proof generation failed")

		if errors.Is(err, badger.ErrKeyNotFound) {
			if removeErr := p.io.DeleteFile(merkle, owner, start); removeErr != nil { // delete the key upon failure?
				log.Error().
					Err(removeErr).
					Msg("Failed to cleanup orphaned file entry")
			}
		}

		return err
	}

	if proof == nil || item == nil {
		return nil
	}

	log.Debug().Msg("Successfully generated proof")

	msg := types.NewMsgPostProof(p.wallet.AccAddress(), merkle, owner, start, item, proof, index)

	m, wg := p.q.Add(msg)

	if m.Index() == -1 { // message was skipped because it was a duplicate
		return nil
	}

	wg.Wait()

	if m.Error() != nil {
		log.Warn().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Err(m.Error()).
			Msg("Proof posting failed, will try again")
		return m.Error()
	}

	if m.Res() == nil {
		log.Warn().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Msg("Message response was nil")
		return nil
	}

	if m.Res().Code != 0 {
		log.Warn().
			Hex("merkle", merkle).
			Str("owner", owner).
			Uint32("code", m.Res().Code).
			Int64("start", start).
			Msgf("response was %s", m.Res().RawLog)
		return nil
	}

	var postRes types.MsgPostProofResponse
	data, err := hex.DecodeString(m.Res().Data)
	if err != nil {
		log.Warn().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Err(err).
			Msg("Could not decode response body")
		return err
	}

	encodingCfg := canine.MakeEncodingConfig()
	var txMsgData sdk.TxMsgData
	err = encodingCfg.Marshaler.Unmarshal(data, &txMsgData)
	if err != nil {
		log.Warn().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Err(err).
			Msg("Could not parse response body")

		return err

	}

	if len(txMsgData.Data) == 0 {
		log.Debug().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Msg("No response data")
		return nil
	}

	err = postRes.Unmarshal(txMsgData.Data[m.Index()].Data)
	if err != nil {
		log.Warn().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Err(err).
			Msg("Could not unmarshal response body")

		return err
	}

	if !postRes.Success {
		log.Warn().
			Hex("merkle", merkle).
			Str("owner", owner).
			Int64("start", start).
			Err(errors.New(postRes.ErrorMessage)).
			Msg("Failed to prove file")
	}

	return nil
}

func (p *Prover) Start() {
	p.running = true
	for p.running {
		if !p.running {
			return
		}

		time.Sleep(time.Millisecond * 1000)                                                  // pauses for one second
		if !p.processed.Add(time.Second * time.Duration(p.interval+10)).Before(time.Now()) { // 10 seconds plus the interval
			continue
		}

		if !p.processed.Add(time.Minute*30).Before(time.Now()) &&
			p.lastCount > 0 &&
			p.q.Count() > p.lastCount { // 30 mins have not yet passed, so we check the queue size

			// don't run if the queue has more than the amount of files on disk
			log.Warn().
				Msg("Queue is full, skipping proof cycle")
			continue

		}

		log.Debug().Msg("Starting proof cycle...")

		c := context.Background()
		abciInfo, err := p.wallet.Client.RPCClient.ABCIInfo(c)
		if err != nil {
			log.Error().Err(err)
			continue
		}
		height := abciInfo.Response.LastBlockHeight

		limit := 5000
		unconfirmedTxs, err := p.wallet.Client.RPCClient.UnconfirmedTxs(c, &limit)
		if err != nil {
			log.Error().Err(err).Msg("could not get mempool status")
			return
		}
		if unconfirmedTxs.Total > 2000 {
			log.Error().Msg("Cannot make proofs when mempool is too large.")
			return
		}

		var count int // reset last count here
		t := time.Now()

		err = p.io.ProcessFiles(func(merkle []byte, owner string, start int64) {
			for p.Full() {
				log.Debug().Msg("Proving queue is full, waiting...")

				time.Sleep(time.Second * 5)
			}
			log.Debug().Msg(fmt.Sprintf("proving: %x", merkle))
			filesProving.Inc()
			p.Inc()
			count++
			go p.wrapPostProof(merkle, owner, start, height, t)
		})
		if err != nil {
			log.Error().Err(err)
		}

		p.lastCount = count

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
		// disable deleting the file if it's not ours

		//if err.Error() == ErrNotOurs { // if the file is not ours, delete it
		//	log.Debug().
		//		Hex("merkle", merkle).
		//		Str("owner", owner).
		//		Int64("start", start).
		//		Msg("deleting the file that does not belong to this provider")
		//
		//	err := p.io.DeleteFile(merkle, owner, start)
		//	if err != nil {
		//		log.Error().
		//			Hex("merkle", merkle).
		//			Err(err).
		//			Msg("failed to delete file that does not belong to this provider")
		//	}
		//}
	}
}

func (p *Prover) Stop() {
	p.running = false
}

func NewProver(wallet *wallet.Wallet, q *queue.Queue, io FileSystem, interval uint64, threads int16, chunkSize int) *Prover {
	p := Prover{
		running:   false,
		wallet:    wallet,
		q:         q,
		processed: time.Time{},
		interval:  interval,
		io:        io,
		threads:   threads,
		chunkSize: chunkSize,
	}

	return &p
}
