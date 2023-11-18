package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/JackalLabs/sequoia/file_system"

	"github.com/JackalLabs/sequoia/monitoring"

	"github.com/cosmos/gogoproto/grpc"

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
)

type App struct {
	api          *api.API
	q            *queue.Queue
	prover       *proofs.Prover
	strayManager *strays.StrayManager
	home         string
	monitor      *monitoring.Monitor
	fileSystem   *file_system.FileSystem
}

func NewApp(home string) *App {
	cfg, err := config.Init(home)
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

	apiServer := api.NewAPI(cfg.APICfg.Port)

	f := file_system.NewFileSystem(db)

	return &App{
		fileSystem: f,
		api:        apiServer,
		home:       home,
	}
}

func initProviderOnChain(wallet *wallet.Wallet, ip string, totalSpace int64) error {
	init := storageTypes.NewMsgInitProvider(wallet.AccAddress(), ip, totalSpace, "")

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

func updateSpace(wallet *wallet.Wallet, totalSpace int64) error {
	init := storageTypes.NewMsgSetProviderTotalSpace(wallet.AccAddress(), totalSpace)

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

func updateIp(wallet *wallet.Wallet, ip string) error {
	init := storageTypes.NewMsgSetProviderIP(wallet.AccAddress(), ip)

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

func (a *App) GetStorageParams(client grpc.ClientConn) (storageTypes.Params, error) {
	queryParams := &storageTypes.QueryParams{}

	cl := storageTypes.NewQueryClient(client)

	res, err := cl.Params(context.Background(), queryParams)
	if err != nil {
		return storageTypes.Params{}, err
	}

	return res.Params, nil
}

func (a *App) Start() {
	cfg, err := config.Init(a.home)
	if err != nil {
		panic(err)
	}

	w, err := config.InitWallet(a.home)
	if err != nil {
		panic(err)
	}

	myAddress := w.AccAddress()

	queryParams := &storageTypes.QueryProvider{
		Address: myAddress,
	}

	cl := storageTypes.NewQueryClient(w.Client.GRPCConn)

	claimers := make([]string, 0)

	res, err := cl.Provider(context.Background(), queryParams)
	if err != nil {
		log.Info().Msg("Provider does not exist on network or is not connected...")
		err := initProviderOnChain(w, cfg.Ip, cfg.TotalSpace)
		if err != nil {
			panic(err)
		}
	} else {
		claimers = res.Provider.AuthClaimers

		totalSpace, err := strconv.ParseInt(res.Provider.Totalspace, 10, 64)
		if err != nil {
			if err != nil {
				panic(err)
			}
		}
		if totalSpace != cfg.TotalSpace {
			err := updateSpace(w, cfg.TotalSpace)
			if err != nil {
				panic(err)
			}
		}
		if res.Provider.Ip != cfg.Ip {
			err := updateIp(w, cfg.Ip)
			if err != nil {
				panic(err)
			}
		}
	}

	params, err := a.GetStorageParams(w.Client.GRPCConn)
	if err != nil {
		panic(err)
	}

	myUrl := cfg.Ip

	log.Info().Msg(fmt.Sprintf("Provider started as: %s", myAddress))

	a.q = queue.NewQueue(w, cfg.QueueInterval)
	go a.q.Listen()

	a.prover = proofs.NewProver(w, a.q, a.fileSystem, cfg.ProofInterval)
	a.strayManager = strays.NewStrayManager(w, a.q, cfg.StrayManagerCfg.CheckInterval, cfg.StrayManagerCfg.RefreshInterval, cfg.StrayManagerCfg.HandCount, claimers)
	a.monitor = monitoring.NewMonitor(w)

	// Starting the 4 concurrent services
	go a.api.Serve(a.fileSystem, a.prover, w, params.ChunkSize)
	go a.prover.Start()
	go a.strayManager.Start(a.fileSystem, myUrl, params.ChunkSize)
	go a.monitor.Start()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done // Will block here until user hits ctrl+c

	fmt.Println("Shutting Sequoia down safely...")

	_ = a.api.Close()
	a.q.Stop()
	a.prover.Stop()
	a.strayManager.Stop()
	a.monitor.Stop()

	time.Sleep(time.Second * 30) // give the program some time to shut down
	a.fileSystem.Close()
}
