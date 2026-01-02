package strays

import (
	"math/rand"
	"time"

	"github.com/JackalLabs/sequoia/rpc"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v5/x/storage/types"
)

type Hand struct {
	wallet  *wallet.Wallet
	offset  byte
	stray   *types.UnifiedFile
	running bool
}

type StrayManager struct {
	strays          []*types.UnifiedFile
	wallet          *rpc.FailoverClient
	lastSize        uint64
	rand            *rand.Rand
	interval        time.Duration
	refreshInterval time.Duration
	running         bool
	hands           []*Hand
	processed       time.Time
	refreshed       time.Time
}
