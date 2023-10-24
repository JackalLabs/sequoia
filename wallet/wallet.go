package wallet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/client"
	"github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	canine "github.com/jackalLabs/canine-chain/v3/app"
)

func CreateWallet(seed string, derivation string, chainCfg types.ChainConfig) (*wallet.Wallet, error) {
	accountCfg := types.AccountConfig{
		Mnemonic: seed,       // forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm
		HDPath:   derivation, // m/44'/118'/0'/0/0
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
