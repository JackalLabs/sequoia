package strays

import (
	"time"

	"github.com/JackalLabs/sequoia/utils"

	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/proofs"

	"github.com/JackalLabs/sequoia/network"
	"github.com/JackalLabs/sequoia/queue"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (h *Hand) Stop() {
	h.running = false
}

func (h *Hand) Start(f *file_system.FileSystem, wallet *wallet.Wallet, myUrl string, chunkSize int64) {
	h.running = true
	defer log.Info().Msg("Hand stopped")
	for h.running {
		if !h.running {
			return
		}

		if h.stray == nil {
			time.Sleep(time.Millisecond * 333)
			continue
		}

		signee := h.stray.Owner
		merkle := h.stray.Merkle
		start := h.stray.Start
		proofType := h.stray.ProofType

		err := network.DownloadFile(f, merkle, signee, start, wallet, h.stray.FileSize, myUrl, chunkSize, proofType, utils.GetIPFSParams(h.stray))
		if err != nil {
			log.Error().Err(err)
			h.stray = nil
			continue
		}

		tree, chunk, err := f.GetFileTreeByChunk(merkle, signee, start, 0, int(chunkSize), proofType)
		if err != nil {
			log.Error().Err(err)
			h.stray = nil
			continue
		}

		_, proof, err := proofs.GenerateMerkleProof(tree, 0, chunk)
		if err != nil {
			log.Error().Err(err)
			h.stray = nil
			continue
		}

		jproof, err := json.Marshal(*proof)
		if err != nil {
			log.Error().Err(err)
			h.stray = nil
			continue
		}

		msg := &types.MsgPostProofFor{
			Creator:  h.Address(),
			Item:     chunk,
			HashList: jproof,
			Merkle:   merkle,
			Owner:    signee,
			Start:    start,
			Provider: wallet.AccAddress(),
			ToProve:  0,
		}

		data := walletTypes.NewTransactionData(
			msg,
		).WithGasAuto().WithFeeAuto()

		res, err := h.wallet.BroadcastTxCommit(data)
		if err != nil {
			log.Error().Err(err)
			h.stray = nil
			continue
		}

		if res != nil {
			if res.Code > 0 {
				log.Info().Msg(res.RawLog)
			}
		}

		h.stray = nil

	}
}

func (h *Hand) Address() string {
	return h.wallet.AccAddress()
}

func (h *Hand) Busy() bool {
	return h.stray != nil
}

func (h *Hand) Take(stray *types.UnifiedFile) {
	h.stray = stray
}

func (s *StrayManager) NewHand(q *queue.Queue) (*Hand, error) {
	offset := byte(len(s.hands)) + 1

	w, err := s.wallet.CloneWalletOffset(offset)
	if err != nil {
		return nil, err
	}

	h := &Hand{
		offset: offset,
		wallet: w,
		stray:  nil,
	}

	s.hands = append(s.hands, h)
	return h, nil
}
