package config

import "github.com/desmos-labs/cosmos-go-wallet/types"

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
}

type StrayManagerConfig struct {
	CheckInterval   int64 `yaml:"check_interval"`
	RefreshInterval int64 `yaml:"refresh_interval"`
	HandCount       int   `yaml:"hands"`
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
			RPCAddr:       "https://jackal-testnet-rpc.polkachu.com:443",
			GRPCAddr:      "jackal-testnet-grpc.polkachu.com:17590",
			GasPrice:      "0.02ujkl",
			GasAdjustment: 1.5,
			Bech32Prefix:  "jkl",
		},
		Ip:            "https://example.com",
		TotalSpace:    1092616192, // 1 gib default
		DataDirectory: "$HOME/.sequoia/data",
	}
}
