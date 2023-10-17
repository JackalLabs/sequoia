package core

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/client"
	"github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	canine "github.com/jackalLabs/canine-chain/v3/app"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"os"
	"os/signal"
	"sequoia/api"
	"sequoia/proofs"
	"sequoia/queue"
	"sequoia/strays"
	"syscall"
)

func CreateWallet() (*wallet.Wallet, error) {
	chainCfg := types.ChainConfig{
		Bech32Prefix:  "jkl",
		RPCAddr:       "https://jackal-testnet-rpc.polkachu.com:443",
		GRPCAddr:      "jackal-testnet-grpc.polkachu.com:17590",
		GasPrice:      "0.02ujkl",
		GasAdjustment: 1.5,
	}
	accountCfg := types.AccountConfig{
		Mnemonic: "forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm",
		HDPath:   "m/44'/118'/0'/0/0",
	}

	// Set up the SDK config with the proper bech32 prefixes
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(chainCfg.Bech32Prefix, fmt.Sprintf("%spub", chainCfg.Bech32Prefix))

	encodingCfg := canine.MakeEncodingConfig()

	c, err := client.NewClient(&chainCfg, encodingCfg.Marshaler)
	if err != nil {
		panic(err)
	}

	w, err := wallet.NewWallet(&accountCfg, c, encodingCfg.TxConfig)
	if err != nil {
		panic(err)
	}

	return w, err
}

type App struct {
	db           *badger.DB
	api          *api.API
	q            *queue.Queue
	prover       *proofs.Prover
	strayManager *strays.StrayManager
}

func NewApp() *App {
	db, err := badger.Open(badger.DefaultOptions("data"))
	if err != nil {
		panic(err)
	}

	apiServer := api.NewAPI(3333)

	return &App{
		db:  db,
		api: apiServer,
	}
}

func (a *App) Start() {
	w, err := CreateWallet()
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
		fmt.Println("Provider does not exist on network or is not connected.")
		return
	}

	myUrl := res.Providers.Ip

	fmt.Printf("Provider started as: %s\n", myAddress)

	a.q = queue.NewQueue(w)
	a.prover = proofs.NewProver(w, a.db, a.q)

	go a.api.Serve(a.db, a.q, w)
	go a.prover.Start()
	go a.q.Listen()

	a.strayManager = strays.NewStrayManager(w, a.q, 30, 60, 1, res.Providers.AuthClaimers)

	go a.strayManager.Start(a.db, myUrl)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Press ctrl+c to quit...")
	<-done // Will block here until user hits ctrl+c

	a.db.Close()
	a.q.Stop()
	a.prover.Stop()
	a.strayManager.Stop()
}
