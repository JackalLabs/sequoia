package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/jackalLabs/canine-chain/v4/types/module/testutil"
	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"
)

func GetTxDecoder() sdk.TxDecoder {
	cfg := MakeTestEncodingConfig()

	return cfg.TxConfig.TxDecoder()
}

func MakeTestEncodingConfig() testutil.TestEncodingConfig {
	cfg := testutil.MakeTestEncodingConfig()

	storageTypes.RegisterInterfaces(cfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(cfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feegrant.RegisterInterfaces(cfg.InterfaceRegistry)
	return cfg
}
