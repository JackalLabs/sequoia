package core

import (
	"context"
	"fmt"
	"github.com/JackalLabs/sequoia/api"
	"github.com/JackalLabs/sequoia/config"
	"github.com/JackalLabs/sequoia/logger"
	"github.com/JackalLabs/sequoia/proofs"
	"github.com/JackalLabs/sequoia/queue"
	"github.com/JackalLabs/sequoia/strays"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	db           *badger.DB
	api          *api.API
	q            *queue.Queue
	prover       *proofs.Prover
	strayManager *strays.StrayManager
}

func NewApp() *App {

	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	dataDir := os.ExpandEnv(cfg.DataDirectory)

	err = os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	options := badger.DefaultOptions(dataDir)

	options.Logger = &logger.SequoiaLogger{}
	options.BlockCacheSize = 256 << 25

	db, err := badger.Open(options)
	if err != nil {
		panic(err)
	}

	apiServer := api.NewAPI(3333)

	return &App{
		db:  db,
		api: apiServer,
	}
}

func initProviderOnChain(wallet *wallet.Wallet, ip string, totalSpace int64) error {
	init := storageTypes.NewMsgInitProvider(wallet.AccAddress(), ip, fmt.Sprintf("%d", totalSpace), "")

	data := walletTypes.NewTransactionData(
		init,
	).WithGasAuto().WithFeeAuto()

	builder, err := wallet.BuildTx(data)
	if err != nil {
		return err
	}

	res, err := wallet.Client.BroadcastTxCommit(builder.GetTx())
	if err != nil {
		return err
	}

	log.Info().Msg(res.TxHash)

	return nil
}

func (a *App) Start() {

	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	w, err := config.InitWallet()
	if err != nil {
		panic(err)
	}

	myAddress := w.AccAddress()

	queryParams := &storageTypes.QueryProviderRequest{
		Address: myAddress,
	}

	cl := storageTypes.NewQueryClient(w.Client.GRPCConn)

	res, err := cl.Providers(context.Background(), queryParams)
	if err != nil {
		log.Info().Msg("Provider does not exist on network or is not connected...")
		err := initProviderOnChain(w, cfg.Ip, cfg.TotalSpace)
		if err != nil {
			panic(err)
		}
	}

	myUrl := res.Providers.Ip

	log.Info().Msg(fmt.Sprintf("Provider started as: %s", myAddress))

	a.q = queue.NewQueue(w, cfg.QueueInterval)
	a.prover = proofs.NewProver(w, a.db, a.q, cfg.ProofInterval)

	go a.api.Serve(a.db, a.q, w)
	go a.prover.Start()
	go a.q.Listen()

	a.strayManager = strays.NewStrayManager(w, a.q, cfg.StrayManagerCfg.CheckInterval, cfg.StrayManagerCfg.RefreshInterval, cfg.StrayManagerCfg.HandCount, res.Providers.AuthClaimers)

	go a.strayManager.Start(a.db, myUrl)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done // Will block here until user hits ctrl+c

	fmt.Println("Shutting down safely...")

	a.q.Stop()
	a.prover.Stop()
	a.strayManager.Stop()

	time.Sleep(time.Second * 30) // give the program some time to shut down
	a.db.Close()

}
