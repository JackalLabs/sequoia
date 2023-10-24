package strays

import (
	"math/rand"
	"time"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
)

type Hand struct {
	wallet  *wallet.Wallet
	offset  byte
	stray   *types.Strays
	running bool
}

type StrayManager struct {
	strays          []*types.Strays
	wallet          *wallet.Wallet
	lastSize        int
	rand            *rand.Rand
	interval        time.Duration
	refreshInterval time.Duration
	running         bool
	hands           []*Hand
	processed       time.Time
	refreshed       time.Time
}
