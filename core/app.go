package core

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/client"
	"github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v2"
	"sequoia/api"
	"sequoia/queue"
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

	encodingConfig := MakeEncodingConfig()

	c, err := client.NewClient(&chainCfg, encodingConfig.Marshaler)
	if err != nil {
		panic(err)
	}

	w, err := wallet.NewWallet(&accountCfg, c, encodingConfig.TxConfig)
	if err != nil {
		panic(err)
	}

	return w, err
}

type App struct {
	db  *badger.DB
	api *api.API
	q   *queue.Queue
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

	fmt.Printf("Provider started as: %s", myAddress)

	q := queue.NewQueue(w)
	a.q = q

	a.api.Serve(a.db, q, myAddress)

	defer a.db.Close()
	defer q.Stop()
}
