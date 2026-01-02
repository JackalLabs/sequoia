package strays

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/rpc"

	"github.com/JackalLabs/sequoia/queue"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/jackalLabs/canine-chain/v5/x/storage/types"
	"github.com/rs/zerolog/log"
)

// refreshIntervalBufferSeconds adds a small buffer (in seconds) to the configured
// refresh interval to account for scheduling jitter, network latency, and block
// timing variance so we don't hammer the endpoint exactly on the boundary.
const (
	refreshIntervalBufferSeconds int64  = 15
	maxReturn                    uint64 = 500
)

// NewStrayManager creates and initializes a new StrayManager with the specified number of hands, authorizing each hand to transact on behalf of the provided wallet if not already authorized.
func NewStrayManager(w *rpc.FailoverClient, q *queue.Queue, interval int64, refreshInterval int64, handCount int, authList []string) *StrayManager {
	s := &StrayManager{
		rand:            rand.New(rand.NewSource(time.Now().Unix())),
		wallet:          w,
		interval:        time.Duration(interval),
		running:         false,
		hands:           make([]*Hand, 0),
		processed:       time.Time{},
		refreshed:       time.Time{},
		refreshInterval: time.Duration(refreshInterval + refreshIntervalBufferSeconds),
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

func (s *StrayManager) Start(f *file_system.FileSystem, q *queue.Queue, myUrl string, chunkSize int64) {
	s.running = true
	defer log.Info().Msg("StrayManager stopped")

	for _, hand := range s.hands {
		go hand.Start(f, s.wallet.Wallet(), q, myUrl, chunkSize)
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
	log.Debug().Msg("Refreshing stray list...")

	var val uint64
	reverse := false
	if s.lastSize > maxReturn {
		val = uint64(s.rand.Int63n(int64(s.lastSize)))
		reverse = s.rand.Intn(2) == 0
	}

	page := &query.PageRequest{ // more randomly pick from the stray pile
		Offset:     val,
		Limit:      maxReturn,
		Reverse:    reverse,
		CountTotal: true,
	}

	queryParams := &types.QueryOpenFiles{
		ProviderAddress: s.wallet.AccAddress(),
		Pagination:      page,
	}

	cl := types.NewQueryClient(s.wallet.GRPCConn())

	res, err := cl.OpenFiles(context.Background(), queryParams)
	if err != nil {
		if rpc.IsConnectionError(err) {
			log.Warn().Err(err).Msg("Connection error during stray refresh, attempting failover")
			s.wallet.Failover()
		}
		return err
	}

	strayCount := len(res.Files)
	s.lastSize = res.Pagination.Total
	if strayCount > 0 {
		log.Info().Msgf("Got updated list of strays of size %d", strayCount)
		newStrays := make([]*types.UnifiedFile, strayCount)
		for i := 0; i < strayCount; i++ {
			stray := res.Files[i]
			newStrays[i] = &stray
		}
		s.strays = newStrays
	}

	return nil
}
