package monitoring

import (
	"context"
	"strconv"
	"time"

	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
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

func (m *Monitor) Start() {
	m.running = true

	for m.running {
		time.Sleep(time.Second * 5)
		m.updateBurns()
		m.updateHeight()
	}
}

func (m *Monitor) Stop() {
	m.running = false
}
