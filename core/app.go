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
	"github.com/JackalLabs/sequoia/ipfs"
	"github.com/ipfs/boxo/blockstore"

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
	badger "github.com/dgraph-io/badger/v4"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/JackalLabs/sequoia/recycle"
)

type App struct {
	api          *api.API
	q            *queue.Queue
	prover       *proofs.Prover
	strayManager *strays.StrayManager
	home         string
	monitor      *monitoring.Monitor
	fileSystem   *file_system.FileSystem
	logFile      *os.File
}

func NewApp(home string) (*App, error) {
	app := App{
		home: home,
	}

	cfg, err := config.Init(home)
	if err != nil {
		return nil, err
	}

	// setup logger to use log file
	if cfg.LogFile != "" {
		path := os.ExpandEnv(cfg.LogFile)

		logf, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0755)
		if err != nil {
			return nil, err
		}

		app.logFile = logf

		stdw := zerolog.ConsoleWriter{Out: os.Stderr}
		fw := zerolog.ConsoleWriter{Out: logf, NoColor: true}
		multi := zerolog.MultiLevelWriter(stdw, fw)
		log.Logger = log.Output(multi)
	}

	ctx := context.Background()

	dataDir := os.ExpandEnv(cfg.DataDirectory)

	err = os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	options := badger.DefaultOptions(dataDir)

	options.Logger = &logger.SequoiaLogger{}
	options.BlockCacheSize = 256 << 25
	options.MaxLevels = 8

	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}

	ds, err := ipfs.NewBadgerDataStore(db)
	if err != nil {
		return nil, err
	}

	bsDir := os.ExpandEnv(cfg.BlockStoreConfig.Directory)
	var bs blockstore.Blockstore
	bs = nil
	switch cfg.BlockStoreConfig.Type {
	case config.OptBadgerDS:
	case config.OptFlatFS:
		bs, err = ipfs.NewFlatfsBlockStore(bsDir)
		if err != nil {
			return nil, err
		}
	}

	apiServer := api.NewAPI(cfg.APICfg.Port)

	f, err := file_system.NewFileSystem(ctx, db, ds, bs, cfg.APICfg.IPFSPort, cfg.APICfg.IPFSDomain)
	if err != nil {
		return nil, err
	}

	app.fileSystem = f
	app.api = apiServer

	return &app, nil
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

func (a *App) Start() error {
	cfg, err := config.Init(a.home)
	if err != nil {
		return err
	}
	log.Debug().Object("config", cfg).Msg("sequoia config")

	w, err := config.InitWallet(a.home)
	if err != nil {
		return err
	}
	log.Info().Str("provider_address", w.AccAddress()).Send()

	myAddress := w.AccAddress()

	queryParams := &storageTypes.QueryProvider{
		Address: myAddress,
	}

	cl := storageTypes.NewQueryClient(w.Client.GRPCConn)

	claimers := make([]string, 0)

	res, err := cl.Provider(context.Background(), queryParams)
	if err != nil {
		log.Info().Err(err).Msg("Provider does not exist on network or is not connected...")
		err := initProviderOnChain(w, cfg.Ip, cfg.TotalSpace)
		if err != nil {
			return err
		}
	} else {
		log.Debug().
			Str("address", res.Provider.Address).
			Str("ip", res.Provider.Ip).
			Str("totalspace", res.Provider.Totalspace).
			Str("burned_contracts", res.Provider.BurnedContracts).
			Str("keybase_identity", res.Provider.KeybaseIdentity).
			Msg("provider query result")
		claimers = res.Provider.AuthClaimers

		totalSpace, err := strconv.ParseInt(res.Provider.Totalspace, 10, 64)
		if err != nil {
			return err
		}
		if totalSpace != cfg.TotalSpace {
			err := updateSpace(w, cfg.TotalSpace)
			if err != nil {
				return err
			}
		}
		if res.Provider.Ip != cfg.Ip {
			err := updateIp(w, cfg.Ip)
			if err != nil {
				return err
			}
		}
	}

	params, err := a.GetStorageParams(w.Client.GRPCConn)
	if err != nil {
		return err
	}

	a.q = queue.NewQueue(w, cfg.QueueInterval)
	go a.q.Listen()

	prover := proofs.NewProver(w, a.q, a.fileSystem, cfg.ProofInterval, cfg.ProofThreads, int(params.ChunkSize))

	recycleDepot, err := recycle.NewRecycleDepot(
		a.home,
		myAddress,
		params.ChunkSize,
		a.fileSystem,
		prover,
		types.NewQueryClient(w.Client.GRPCConn),
	)
	if err != nil {
		return err
	}

	myUrl := cfg.Ip

	log.Info().Msg(fmt.Sprintf("Provider started as: %s", myAddress))

	a.prover = prover
	a.strayManager = strays.NewStrayManager(w, a.q, cfg.StrayManagerCfg.CheckInterval, cfg.StrayManagerCfg.RefreshInterval, cfg.StrayManagerCfg.HandCount, claimers)
	a.monitor = monitoring.NewMonitor(w)

	logFileName := ""
	if a.logFile != nil {
		logFileName = a.logFile.Name()
	}
	// Starting the 4 concurrent services
	// nolint:all
	go a.api.Serve(recycleDepot, a.fileSystem, a.prover, w, params.ChunkSize, logFileName)
	go a.prover.Start()
	go a.strayManager.Start(a.fileSystem, myUrl, params.ChunkSize)
	go a.monitor.Start()
	go recycleDepot.Start(cfg.StrayManagerCfg.CheckInterval)

	done := make(chan os.Signal, 1)
	defer signal.Stop(done) // undo signal.Notify effect

	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done // Will block here until user hits ctrl+c

	fmt.Println("Shutting Sequoia down safely...")

	recycleDepot.Stop()
	_ = a.api.Close()
	a.q.Stop()
	a.prover.Stop()
	a.strayManager.Stop()
	a.monitor.Stop()

	time.Sleep(time.Second * 30) // give the program some time to shut down
	a.fileSystem.Close()

	return nil
}

func (a *App) Salvage(jprovdHome string) error {
	cfg, err := config.Init(a.home)
	if err != nil {
		return err
	}

	w, err := config.InitWallet(a.home)
	if err != nil {
		return err
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
			return err
		}
	} else {
		claimers = res.Provider.AuthClaimers

		totalSpace, err := strconv.ParseInt(res.Provider.Totalspace, 10, 64)
		if err != nil {
			return err
		}
		if totalSpace != cfg.TotalSpace {
			err := updateSpace(w, cfg.TotalSpace)
			if err != nil {
				return err
			}
		}
		if res.Provider.Ip != cfg.Ip {
			err := updateIp(w, cfg.Ip)
			if err != nil {
				return err
			}
		}
	}

	params, err := a.GetStorageParams(w.Client.GRPCConn)
	if err != nil {
		return err
	}

	a.q = queue.NewQueue(w, cfg.QueueInterval)
	go a.q.Listen()

	prover := proofs.NewProver(w, a.q, a.fileSystem, cfg.ProofInterval, cfg.ProofThreads, int(params.ChunkSize))

	recycleDepot, err := recycle.NewRecycleDepot(
		a.home,
		myAddress,
		params.ChunkSize,
		a.fileSystem,
		prover,
		types.NewQueryClient(w.Client.GRPCConn),
	)
	if err != nil {
		return err
	}

	myUrl := cfg.Ip

	log.Info().Msg(fmt.Sprintf("Provider started as: %s", myAddress))

	a.prover = prover
	a.strayManager = strays.NewStrayManager(w, a.q, cfg.StrayManagerCfg.CheckInterval, cfg.StrayManagerCfg.RefreshInterval, cfg.StrayManagerCfg.HandCount, claimers)
	a.monitor = monitoring.NewMonitor(w)

	done := make(chan os.Signal, 1)
	defer signal.Stop(done) // undo signal.Notify effect

	logFileName := ""
	if a.logFile != nil {
		logFileName = a.logFile.Name()
	}
	// Starting the 4 concurrent services
	// nolint:all
	go a.api.Serve(recycleDepot, a.fileSystem, a.prover, w, params.ChunkSize, logFileName)
	go a.prover.Start()
	go a.strayManager.Start(a.fileSystem, myUrl, params.ChunkSize)
	go a.monitor.Start()
	go recycleDepot.Start(cfg.StrayManagerCfg.CheckInterval)

	err = recycleDepot.SalvageFiles(jprovdHome)
	if err != nil {
		return err
	}

	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done // Will block here until user hits ctrl+c

	fmt.Println("Shutting Sequoia down safely...")

	_ = a.api.Close()
	a.q.Stop()
	a.prover.Stop()
	a.strayManager.Stop()
	a.monitor.Stop()
	recycleDepot.Stop()
	time.Sleep(time.Second * 30) // give the program some time to shut down
	a.fileSystem.Close()

	if a.logFile != nil {
		return a.logFile.Close()
	}

	return nil
}
