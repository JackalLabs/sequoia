package strays

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
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

// NewStrayManager creates and initializes a new StrayManager with the specified number of hands, authorizing each hand to transact on behalf of the provided wallet if not already authorized.
func NewStrayManager(w *wallet.Wallet, queryClient types.QueryClient, q queue.Queue, interval int64, refreshInterval int64, handWallets []*wallet.Wallet) (*StrayManager, error) {
	s := &StrayManager{
		rand:            rand.New(rand.NewSource(time.Now().Unix())),
		wallet:          w,
		interval:        time.Duration(interval),
		running:         false,
		hands:           make([]*Hand, 0),
		processed:       time.Time{},
		refreshed:       time.Time{},
		refreshInterval: time.Duration(refreshInterval),
		queryClient:     queryClient,
	}

	query := &types.QueryProvider{
		Address: w.AccAddress(),
	}

	res, err := queryClient.Provider(context.Background(), query)
	if err != nil {
		return nil, errors.Join(errors.New("unable to query provider auth claimers"), err)
	}

	authList := res.Provider.AuthClaimers

	for i, wallet := range handWallets {
		log.Info().Msg(fmt.Sprintf("Authorizing hand %d to transact on my behalf...", i))

		h, err := s.NewHand(q)
		if err != nil {
			log.Error().Err(err).Int("index", i).Msg("Failed to create hand")
			return nil, err
		}

		if slices.Contains(authList, wallet.AccAddress()) { // already authorized
			continue
		}

		msg := types.NewMsgAddClaimer(w.AccAddress(), h.Address())

		allowance := feegrant.BasicAllowance{
			SpendLimit: nil,
			Expiration: nil,
		}

		wadd, err := sdk.AccAddressFromBech32(w.AccAddress())
		if err != nil {
			return nil, err
		}

		hadd, err := sdk.AccAddressFromBech32(h.Address())
		if err != nil {
			return nil, err
		}

		grantMsg, nerr := feegrant.NewMsgGrantAllowance(&allowance, wadd, hadd)
		if nerr != nil {
			return nil, err
		}

		m, wg := q.Add(msg)
		q.Add(grantMsg)

		wg.Wait()

		if m.Error() != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *StrayManager) Start(f *file_system.FileSystem, queryClient types.QueryClient, q queue.Queue, myUrl string, chunkSize int64) {
	s.running = true
	defer log.Info().Msg("StrayManager stopped")

	for _, hand := range s.hands {
		go hand.Start(f, s.wallet, queryClient, q, myUrl, chunkSize)
	}

	for s.running {
		if !s.running {
			return
		}

		time.Sleep(time.Millisecond * 333)
		if s.refreshed.Add(time.Second * s.refreshInterval).Before(time.Now()) {
			err := s.RefreshList()
			if err != nil {
				log.Error().Err(err).Msg("failed refresh")
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
	reverse := false
	if s.lastSize > 300 {
		val = uint64(s.rand.Int63n(s.lastSize))
		reverse = s.rand.Intn(2) == 0
	}

	page := &query.PageRequest{ // more randomly pick from the stray pile
		Offset:     val,
		Limit:      300,
		Reverse:    reverse,
		CountTotal: true,
	}

	queryParams := &types.QueryOpenFiles{
		ProviderAddress: s.wallet.AccAddress(),
		Pagination:      page,
	}

	res, err := s.queryClient.OpenFiles(context.Background(), queryParams)
	if err != nil {
		return err
	}
	log.Info().Msgf("Got updated list of strays of size %d", len(res.Files))

	for _, stray := range res.Files {
		newStray := stray
		s.strays = append(s.strays, &newStray)
	}

	s.lastSize = int64(res.Pagination.Total)

	return nil
}
