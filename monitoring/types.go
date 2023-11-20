package monitoring

import (
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var fileBurnCount = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "sequoia_burn_count",
	Help: "The number of files burned by provider",
})

var blockHeight = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "sequoia_block_height",
	Help: "The height of the chain at a given time",
})

var catchingUp = promauto.NewSummary(prometheus.SummaryOpts{
	Name: "sequoia_catching_up",
	Help: "If the node is catching up",
})

var tokenBalance = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "sequoia_balance",
	Help: "Provider token balance",
})

var _ = catchingUp

type Monitor struct {
	running bool
	wallet  *wallet.Wallet
}

func NewMonitor(wallet *wallet.Wallet) *Monitor {
	return &Monitor{
		running: false,
		wallet:  wallet,
	}
}
