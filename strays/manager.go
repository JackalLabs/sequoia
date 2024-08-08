package strays

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/JackalLabs/sequoia/file_system"

	"github.com/JackalLabs/sequoia/queue"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)

func NewStrayManager(w *wallet.Wallet, q *queue.Queue, interval int64, refreshInterval int64, handCount int, authList []string) *StrayManager {
	s := &StrayManager{
		rand:            rand.New(rand.NewSource(time.Now().Unix())),
		wallet:          w,
		interval:        time.Duration(interval),
		running:         false,
		hands:           make([]*Hand, 0),
		processed:       time.Time{},
		refreshed:       time.Time{},
		refreshInterval: time.Duration(refreshInterval),
	}

	for i := 0; i < handCount; i++ {
		log.Info().Msg(fmt.Sprintf("Authorizing hand %d to transact on my behalf...", i))

		h, err := s.NewHand(q)
		if err != nil {
			log.Error().Err(err)
			continue
		}

		alreadyAuth := false
		for _, auth := range authList {
			if auth == h.Address() {
				alreadyAuth = true
				break
			}
		}

		if alreadyAuth {
			continue
		}

		msg := types.NewMsgAddClaimer(w.AccAddress(), h.Address())

		allowance := feegrant.BasicAllowance{
			SpendLimit: nil,
			Expiration: nil,
		}

		wadd, err := sdk.AccAddressFromBech32(w.AccAddress())
		if err != nil {
			log.Error().Err(err)
			continue
		}

		hadd, err := sdk.AccAddressFromBech32(h.Address())
		if err != nil {
			log.Error().Err(err)
			continue
		}

		grantMsg, nerr := feegrant.NewMsgGrantAllowance(&allowance, wadd, hadd)
		if nerr != nil {
			log.Error().Err(nerr)
			continue
		}

		m, wg := q.Add(msg)
		q.Add(grantMsg)

		wg.Wait()

		if m.Error() != nil {
			log.Error().Err(m.Error())
		}
	}

	return s
}

func (s *StrayManager) Start(f *file_system.FileSystem, myUrl string, chunkSize int64) {
	s.running = true
	defer log.Info().Msg("StrayManager stopped")

	for _, hand := range s.hands {
		go hand.Start(f, s.wallet, myUrl, chunkSize)
	}

	for s.running {
		if !s.running {
			return
		}

		time.Sleep(time.Millisecond * 333)
		if s.refreshed.Add(time.Second * s.refreshInterval).Before(time.Now()) {
			err := s.RefreshList()
			if err != nil {
				log.Error().Err(err)

				log.Info().Msg("failed refresh")
			}
			s.refreshed = time.Now()
		}

		if !s.processed.Add(time.Second * s.interval).Before(time.Now()) {
			continue
		}

		for _, hand := range s.hands {
			if hand.Busy() {
				continue
			}

			hand.Take(s.Pop())
		}

		s.processed = time.Now()

	}
}

func (s *StrayManager) Pop() *types.UnifiedFile {
	if len(s.strays) == 0 {
		return nil
	}

	stray := s.strays[0]
	s.strays = s.strays[1:]

	return stray
}

func (s *StrayManager) Stop() {
	s.running = false

	for _, hand := range s.hands {
		hand.Stop()
	}
}

func (s *StrayManager) RefreshList() error {
	log.Info().Msg("Refreshing stray list...")

	s.strays = make([]*types.UnifiedFile, 0)

	var val uint64
	if s.lastSize > 300 {
		val = uint64(s.rand.Int63n(int64(s.lastSize)))
	}

	page := &query.PageRequest{ // more randomly pick from the stray pile
		Offset:     val,
		Limit:      300,
		Reverse:    s.rand.Intn(2) == 0,
		CountTotal: true,
	}

	queryParams := &types.QueryOpenFiles{
		ProviderAddress: s.wallet.AccAddress(),
		Pagination:      page,
	}

	cl := types.NewQueryClient(s.wallet.Client.GRPCConn)

	res, err := cl.OpenFiles(context.Background(), queryParams)
	if err != nil {
		return err
	}

	for _, stray := range res.Files {
		newStray := stray
		s.strays = append(s.strays, &newStray)
	}

	s.lastSize = res.Size()

	return nil
}
