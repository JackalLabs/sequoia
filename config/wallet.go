package config

import (
	"fmt"
	"os"
	"path"

	sequoiaWallet "github.com/JackalLabs/sequoia/wallet"
	"github.com/cosmos/go-bip39"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const SeedFileName = "provider_wallet.json"

func generateWallet() (*Seed, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return nil, err
	}

	s := &Seed{
		SeedPhrase:     mnemonic, //"forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm",
		DerivationPath: "m/44'/118'/0'/0/0",
	}

	return s, nil
}

func ImportSeed(seedData []byte) (*Seed, error) {
	seed := Seed{}

	err := json.Unmarshal(seedData, &seed)
	if err != nil {
		return nil, err
	}

	return &seed, nil
}

func (s *Seed) Export() ([]byte, error) {
	return json.Marshal(s)
}

func createWallet(directory string) error {
	wallet, err := generateWallet()
	if err != nil {
		return err
	}

	seedData, err := wallet.Export()
	if err != nil {
		return err
	}

	newWallet, err := createIfNotExists(directory, SeedFileName, seedData)
	if err != nil {
		return err
	}

	filePath := path.Join(directory, SeedFileName)

	if newWallet {
		fmt.Printf("A new wallet was just created with a random seed phrase, if you wish to use an existing seed phrase, edit %s\n", filePath)
		os.Exit(0)
	}

	return nil
}

func InitWallet(home string) (*wallet.Wallet, error) {
	directory := os.ExpandEnv(home)

	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = createWallet(directory)
	if err != nil {
		return nil, err
	}

	seedData, err := readFile(directory, SeedFileName)
	if err != nil {
		return nil, err
	}

	seed, err := ImportSeed(seedData)
	if err != nil {
		return nil, err
	}

	config, err := ReadConfigFile(directory)
	if err != nil {
		return nil, err
	}

	return sequoiaWallet.CreateWallet(seed.SeedPhrase, seed.DerivationPath, config.ChainCfg)
}
