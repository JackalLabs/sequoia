package monitoring

import (
	"context"
	"strconv"
	"time"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)

func (m *Monitor) updateBurns() {
	cl := types.NewQueryClient(m.wallet.Client.GRPCConn)
	provRes, err := cl.Provider(context.Background(), &types.QueryProvider{Address: m.wallet.AccAddress()})
	if err != nil {
		return
	}

	bString := provRes.Provider.BurnedContracts
	burns, err := strconv.ParseInt(bString, 10, 64)
	if err != nil {
		return
	}

	fileBurnCount.Set(float64(burns))
}

func (m *Monitor) updateHeight() {
	abciInfo, err := m.wallet.Client.RPCClient.ABCIInfo(context.Background())
	if err != nil {
		return
	}
	height := abciInfo.Response.LastBlockHeight

	blockHeight.Set(float64(height))
}

func (m *Monitor) updateBalance() {
	cl := bankTypes.NewQueryClient(m.wallet.Client.GRPCConn)
	provRes, err := cl.Balance(context.Background(), &bankTypes.QueryBalanceRequest{Address: m.wallet.AccAddress(), Denom: "ujkl"})
	if err != nil {
		return
	}

	amt := provRes.Balance.Amount

	tokenBalance.Set(float64(amt.QuoRaw(1_000_000).Int64()))
}

func (m *Monitor) Start() {
	defer log.Info().Msg("Monitor moduel stopped")
	m.running = true

	for m.running {
		time.Sleep(time.Second * 5)
		m.updateBurns()
		m.updateHeight()
		m.updateBalance()
	}
}

func (m *Monitor) Stop() {
	m.running = false
}
