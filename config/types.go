package config

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Seed struct {
	SeedPhrase     string `json:"seed_phrase"`
	DerivationPath string `json:"derivation_path"`
}

// required for the mapstructure tag
type ChainConfig struct {
	Bech32Prefix  string  `yaml:"bech32_prefix" mapstructure:"bech32_prefix"`
	RPCAddr       string  `yaml:"rpc_addr" mapstructure:"rpc_addr"`
	GRPCAddr      string  `yaml:"grpc_addr" mapstructure:"grpc_addr"`
	GasPrice      string  `yaml:"gas_price" mapstructure:"gas_price"`
	GasAdjustment float64 `yaml:"gas_adjustment" mapstructure:"gas_adjustment"`
}

type Config struct {
	QueueInterval    int64              `yaml:"queue_interval" mapstructure:"queue_interval"`
	ProofInterval    int64              `yaml:"proof_interval" mapstructure:"proof_interval"`
	StrayManagerCfg  StrayManagerConfig `yaml:"stray_manager" mapstructure:"stray_manager"`
	ChainCfg         ChainConfig        `yaml:"chain_config" mapstructure:"chain_config"`
	Ip               string             `yaml:"domain" mapstructure:"domain"`
	TotalSpace       int64              `yaml:"total_bytes_offered" mapstructure:"total_bytes_offered"`
	DataDirectory    string             `yaml:"data_directory" mapstructure:"data_directory"`
	APICfg           APIConfig          `yaml:"api_config" mapstructure:"api_config"`
	ProofThreads     int16              `yaml:"proof_threads" mapstructure:"proof_threads"`
	BlockStoreConfig BlockStoreConfig   `yaml:"block_store_config" mapstructure:"block_store_config"`
	LogFile          string             `yaml:"log_file" mapstructure:"log_file"`
	SSHConfig        SSHConfig          `yaml:"ssh_config" mapstructure:"ssh_config"`
}

func DefaultLogFile() string {
	return "$HOME/.sequoia/log/sequoia.log"
}

func DefaultQueueInterval() int64 {
	return 10
}

func DefaultProofInterval() int64 {
	return 120
}

func DefaultIP() string {
	return "https://example.com"
}

func DefaultTotalSpace() int64 {
	return 1092616192
}

func DefaultDataDirectory() string {
	return "$HOME/.sequoia/data"
}

func DefaultProofThreads() int16 {
	return 1000
}

type StrayManagerConfig struct {
	CheckInterval   int64 `yaml:"check_interval" mapstructure:"check_interval"`
	RefreshInterval int64 `yaml:"refresh_interval" mapstructure:"refresh_interval"`
	HandCount       int   `yaml:"hands" mapstructure:"hands"`
}

func DefaultStrayManagerConfig() StrayManagerConfig {
	return StrayManagerConfig{
		CheckInterval:   30,
		RefreshInterval: 120,
		HandCount:       2,
	}
}

type APIConfig struct {
	Port       int64  `yaml:"port" mapstructure:"port"`
	IPFSPort   int    `yaml:"ipfs_port" mapstructure:"ipfs_port"`
	IPFSDomain string `yaml:"ipfs_domain" mapstructure:"ipfs_domain"`
}

func DefaultAPIConfig() APIConfig {
	return APIConfig{
		Port:       3333,
		IPFSPort:   4005,
		IPFSDomain: "dns4/ipfs.example.com/tcp/4001",
	}
}

type SSHConfig struct {
	Enable            bool     `yaml:"enable" mapstructure:"enable"`
	Host              string   `yaml:"host" mapstructure:"host"`
	Port              string   `yaml:"port" mapstructure:"port"`
	AuthorizedPubKeys []string `yaml:"authorized_pub_keys" mapstructure:"authorized_pub_keys"`
	HostKeyFile       string   `yaml:"host_key_file" mapstructure:"host_key_file"`
}

func DefaultSSHConfig() SSHConfig {
	return SSHConfig{
		Enable:            false,
		AuthorizedPubKeys: []string{},
		Host:              "localhost",
		Port:              "23234",
		HostKeyFile:       "",
	}
}

const (
	OptBadgerDS = "badgerds"
	OptFlatFS   = "flatfs"
)

type BlockStoreConfig struct {
	// *choosing badgerdb as block store will need to use the same directory
	// for data directory
	Directory string `yaml:"directory" mapstructure:"directory"`
	// data store options: flatfs, badgerdb
	Type string `yaml:"type" mapstructure:"type"`
}

func DefaultBlockStoreConfig() BlockStoreConfig {
	return BlockStoreConfig{
		Directory: "$HOME/.sequoia/blockstore",
		Type:      OptFlatFS,
	}
}

// LegacyWallet handles keys from earlier versions of storage providers.
// v3 and earlier providers used private key to sign txs
// and by design it can't derive mnemonic seed which made
// it incompatible with sequoia's old wallet creation.
type LegacyWallet struct {
	Key     string `json:"key"`
	Address string `json:"address"`
}

func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		RPCAddr:       "http://localhost:26657",
		GRPCAddr:      "localhost:9090",
		GasPrice:      "0.02ujkl",
		GasAdjustment: 1.5,
		Bech32Prefix:  "jkl",
	}
}

func DefaultConfig() *Config {
	return &Config{
		QueueInterval:    DefaultQueueInterval(),
		ProofInterval:    DefaultProofInterval(),
		StrayManagerCfg:  DefaultStrayManagerConfig(),
		ChainCfg:         DefaultChainConfig(),
		Ip:               DefaultIP(),
		TotalSpace:       DefaultTotalSpace(), // 1 gib default
		DataDirectory:    DefaultDataDirectory(),
		APICfg:           DefaultAPIConfig(),
		ProofThreads:     DefaultProofThreads(),
		BlockStoreConfig: DefaultBlockStoreConfig(),
		LogFile:          DefaultLogFile(),
		SSHConfig:        DefaultSSHConfig(),
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
		Int16("ProofThreads", c.ProofThreads).
		Str("BlockstoreBackend", c.BlockStoreConfig.Type)
}

func init() {
	viper.SetDefault("QueueInterval", DefaultQueueInterval())
	viper.SetDefault("ProofInterval", DefaultProofInterval())
	viper.SetDefault("StrayManagerCfg", DefaultStrayManagerConfig())
	viper.SetDefault("ChainCfg", DefaultChainConfig())
	viper.SetDefault("Ip", DefaultIP())
	viper.SetDefault("TotalSpace", DefaultTotalSpace()) // 1 gib defaul
	viper.SetDefault("DataDirectory", DefaultDataDirectory())
	viper.SetDefault("APICfg", DefaultAPIConfig())
	viper.SetDefault("ProofThreads", DefaultProofThreads())
	viper.SetDefault("BlockStoreConfig", DefaultBlockStoreConfig())
	viper.SetDefault("LogFile", DefaultLogFile())
	viper.SetDefault("SSHConfig", DefaultSSHConfig())
}

func (c *Config) expandPaths() {
	c.DataDirectory = expandPath(c.DataDirectory)
	c.BlockStoreConfig.Directory = expandPath(c.BlockStoreConfig.Directory)
	c.LogFile = expandPath(c.LogFile)
	c.SSHConfig.HostKeyFile = expandPath(c.SSHConfig.HostKeyFile)
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		dir, err := os.UserHomeDir()
		if err != nil {
			return path
		}

		return filepath.Join(dir, path[1:])
	}
	return os.ExpandEnv(path)
}
