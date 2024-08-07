package config

import (
	"github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/rs/zerolog"
)

type Seed struct {
	SeedPhrase     string `json:"seed_phrase"`
	DerivationPath string `json:"derivation_path"`
}
type Config struct {
	QueueInterval   int64              `yaml:"queue_interval"`
	ProofInterval   int64              `yaml:"proof_interval"`
	StrayManagerCfg StrayManagerConfig `yaml:"stray_manager"`
	ChainCfg        types.ChainConfig  `yaml:"chain_config"`
	Ip              string             `yaml:"domain"`
	TotalSpace      int64              `yaml:"total_bytes_offered"`
	DataDirectory   string             `yaml:"data_directory"`
	APICfg          APIConfig          `yaml:"api_config"`
	ProofThreads    int64              `yaml:"proof_threads"`
}

type StrayManagerConfig struct {
	CheckInterval   int64 `yaml:"check_interval"`
	RefreshInterval int64 `yaml:"refresh_interval"`
	HandCount       int   `yaml:"hands"`
}

type APIConfig struct {
	Port       int64  `yaml:"port"`
	IPFSPort   int    `yaml:"ipfs_port"`
	IPFSDomain string `yaml:"ipfs_domain"`
}

// LegacyWallet handles keys from earlier versions of storage providers.
// v3 and earlier providers used private key to sign txs
// and by design it can't derive mnemonic seed which made
// it incompatible with sequoia's old wallet creation.
type LegacyWallet struct {
	Key     string `json:"key"`
	Address string `json:"address"`
}

func DefaultConfig() *Config {
	return &Config{
		QueueInterval: 10,
		ProofInterval: 120,
		StrayManagerCfg: StrayManagerConfig{
			CheckInterval:   30,
			RefreshInterval: 120,
			HandCount:       2,
		},
		ChainCfg: types.ChainConfig{
			RPCAddr:       "http://localhost:26657",
			GRPCAddr:      "127.0.0.1:9090",
			GasPrice:      "0.02ujkl",
			GasAdjustment: 1.5,
			Bech32Prefix:  "jkl",
		},
		Ip:            "https://example.com",
		TotalSpace:    1092616192, // 1 gib default
		DataDirectory: "$HOME/.sequoia/data",
		APICfg: APIConfig{
			Port:       3333,
			IPFSPort:   4005,
			IPFSDomain: "dns4/ipfs.example.com/tcp/4001",
		},
		ProofThreads: 1000,
	}
}

func (c Config) MarshalZerologObject(e *zerolog.Event) {
	e.Int64("QueueInterval", c.QueueInterval).
		Int64("ProofInterval", c.ProofInterval).
		Int64("StrayCheckInterval", c.StrayManagerCfg.CheckInterval).
		Int64("StrayRefreshInterval", c.StrayManagerCfg.RefreshInterval).
		Int("StrayHandCount", c.StrayManagerCfg.HandCount).
		Str("ChainRPCAddr", c.ChainCfg.RPCAddr).
		Str("ChainGRPCAddr", c.ChainCfg.GRPCAddr).
		Str("ChainGasPrice", c.ChainCfg.GasPrice).
		Float64("ChainGasAdjustment", c.ChainCfg.GasAdjustment).
		Str("IP", c.Ip).
		Int64("TotalSpace", c.TotalSpace).
		Str("DataDirectory", c.DataDirectory).
		Int64("APIPort", c.APICfg.Port).
		Int("APIIPFSPort", c.APICfg.IPFSPort).
		Str("APIIPFSDomain", c.APICfg.IPFSDomain).
		Int64("ProofThreads", c.ProofThreads)
}
