package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	return cfg
}
