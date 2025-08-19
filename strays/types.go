package strays

import (
	"math/rand"
	"time"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
)

type Hand struct {
	wallet  *wallet.Wallet
	offset  byte
	stray   *types.UnifiedFile
	running bool
}

type StrayManager struct {
	strays          []*types.UnifiedFile
	wallet          *wallet.Wallet
	lastSize        int64
	rand            *rand.Rand
	interval        time.Duration
	refreshInterval time.Duration
	running         bool
	hands           []*Hand
	processed       time.Time
	refreshed       time.Time
	queryClient     types.QueryClient
}
