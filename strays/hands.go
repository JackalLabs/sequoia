package strays

import (
	"strconv"
	"time"

	"github.com/JackalLabs/sequoia/network"
	"github.com/JackalLabs/sequoia/queue"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"
)

func (h *Hand) Stop() {
	h.running = false
}

func (h *Hand) Start(db *badger.DB, wallet *wallet.Wallet, myUrl string) {
	h.running = true
	for h.running {

		if h.stray == nil {
			time.Sleep(time.Millisecond * 333)
			continue
		}

		signee := h.stray.Signee
		fid := h.stray.Fid
		cid := h.stray.Cid
		size, err := strconv.ParseInt(h.stray.Filesize, 10, 64)
		if err != nil {
			log.Error().Err(err)
			h.stray = nil
			continue
		}

		err = network.DownloadFile(db, cid, fid, wallet, signee, size, myUrl)
		if err != nil {
			log.Error().Err(err)
			h.stray = nil
			continue
		}

		msg := &types.MsgClaimStray{
			Creator:    h.Address(),
			Cid:        cid,
			ForAddress: wallet.AccAddress(),
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

func (h *Hand) Take(stray *types.Strays) {
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
