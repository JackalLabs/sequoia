package rpc

import (
	"context"
	"errors"
	"strings"
	"sync"

	sequoiaWallet "github.com/JackalLabs/sequoia/wallet"
	"github.com/cosmos/gogoproto/grpc"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/rpc/client"
)

// ErrNoNodes is returned when no RPC/GRPC nodes are configured.
var ErrNoNodes = errors.New("no RPC/GRPC nodes configured")

// NodeConfig contains the configuration needed to connect to blockchain nodes.
// This is separate from config.ChainConfig to avoid import cycles.
type NodeConfig struct {
	Bech32Prefix  string
	RPCAddrs      []string
	GRPCAddrs     []string
	GasPrice      string
	GasAdjustment float64
}

// FailoverClient wraps a wallet and provides automatic failover between
// multiple RPC/GRPC nodes. When a connection error is detected, it
// automatically switches to the next available node.
type FailoverClient struct {
	mu sync.RWMutex

	wallet *wallet.Wallet

	// Configuration
	nodeCfg      NodeConfig
	seed         string
	derivation   string
	useLegacyKey bool
	legacyKey    string

	// Node tracking
	currentIndex  int
	failoverCount int
}

// NewFailoverClient creates a new FailoverClient with the given configuration.
// It initializes the first connection using the provided seed phrase.
func NewFailoverClient(nodeCfg NodeConfig, seed, derivation string) (*FailoverClient, error) {
	if len(nodeCfg.RPCAddrs) == 0 || len(nodeCfg.GRPCAddrs) == 0 {
		return nil, ErrNoNodes
	}

	fc := &FailoverClient{
		nodeCfg:      nodeCfg,
		seed:         seed,
		derivation:   derivation,
		useLegacyKey: false,
		currentIndex: 0,
	}

	if err := fc.connectToFirstAvailable(); err != nil {
		return nil, err
	}

	return fc, nil
}

// NewFailoverClientWithPrivKey creates a new FailoverClient using a legacy private key.
func NewFailoverClientWithPrivKey(nodeCfg NodeConfig, privKey string) (*FailoverClient, error) {
	if len(nodeCfg.RPCAddrs) == 0 || len(nodeCfg.GRPCAddrs) == 0 {
		return nil, ErrNoNodes
	}

	fc := &FailoverClient{
		nodeCfg:      nodeCfg,
		useLegacyKey: true,
		legacyKey:    privKey,
		currentIndex: 0,
	}

	if err := fc.connectToFirstAvailable(); err != nil {
		return nil, err
	}

	return fc, nil
}

// connectToFirstAvailable attempts to connect to the first available node.
func (fc *FailoverClient) connectToFirstAvailable() error {
	var lastErr error
	for i := 0; i < len(fc.nodeCfg.RPCAddrs); i++ {
		w, err := fc.createWalletAtIndex(i)
		if err == nil {
			fc.wallet = w
			fc.currentIndex = i
			log.Info().
				Int("node_index", fc.currentIndex).
				Str("rpc", fc.nodeCfg.RPCAddrs[fc.currentIndex]).
				Str("grpc", fc.nodeCfg.GRPCAddrs[fc.currentIndex]).
				Msg("Connected to blockchain node")
			return nil
		}
		lastErr = err
		if i > 0 {
			log.Warn().Err(err).Int("index", i).Msg("Failed to connect to node, trying next")
		}
	}
	return lastErr
}

// createWalletAtIndex creates a new wallet connection using the node at the given index.
func (fc *FailoverClient) createWalletAtIndex(index int) (*wallet.Wallet, error) {
	if index >= len(fc.nodeCfg.RPCAddrs) || index >= len(fc.nodeCfg.GRPCAddrs) {
		index = 0
	}

	// Create a modified chain config with the specific node addresses
	chainCfg := walletTypes.ChainConfig{
		Bech32Prefix:  fc.nodeCfg.Bech32Prefix,
		RPCAddr:       fc.nodeCfg.RPCAddrs[index],
		GRPCAddr:      fc.nodeCfg.GRPCAddrs[index],
		GasPrice:      fc.nodeCfg.GasPrice,
		GasAdjustment: fc.nodeCfg.GasAdjustment,
	}

	if fc.useLegacyKey {
		return sequoiaWallet.CreateWalletPrivKey(fc.legacyKey, chainCfg)
	}
	return sequoiaWallet.CreateWallet(fc.seed, fc.derivation, chainCfg)
}

// Failover switches to the next available node. Returns true if a new node
// was connected, false if we've cycled through all nodes without success.
func (fc *FailoverClient) Failover() bool {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	startIndex := fc.currentIndex
	totalNodes := len(fc.nodeCfg.RPCAddrs)

	for i := 1; i <= totalNodes; i++ {
		nextIndex := (startIndex + i) % totalNodes
		log.Info().
			Int("from_index", fc.currentIndex).
			Int("to_index", nextIndex).
			Str("rpc", fc.nodeCfg.RPCAddrs[nextIndex]).
			Str("grpc", fc.nodeCfg.GRPCAddrs[nextIndex]).
			Msg("Attempting failover to next node")

		w, err := fc.createWalletAtIndex(nextIndex)
		if err != nil {
			log.Warn().Err(err).
				Int("index", nextIndex).
				Str("rpc", fc.nodeCfg.RPCAddrs[nextIndex]).
				Msg("Failed to connect during failover, trying next")
			continue
		}

		fc.wallet = w
		fc.currentIndex = nextIndex
		fc.failoverCount++

		log.Info().
			Int("node_index", fc.currentIndex).
			Str("rpc", fc.nodeCfg.RPCAddrs[fc.currentIndex]).
			Str("grpc", fc.nodeCfg.GRPCAddrs[fc.currentIndex]).
			Int("total_failovers", fc.failoverCount).
			Msg("Successfully failed over to new node")

		return true
	}

	log.Error().Msg("Failed to connect to any node during failover")
	return false
}

// Wallet returns the underlying wallet. Use with caution - prefer using
// the FailoverClient methods which handle automatic failover.
func (fc *FailoverClient) Wallet() *wallet.Wallet {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.wallet
}

// GRPCConn returns the current GRPC connection.
func (fc *FailoverClient) GRPCConn() grpc.ClientConn {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.wallet.Client.GRPCConn
}

// RPCClient returns the current RPC client.
func (fc *FailoverClient) RPCClient() client.Client {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.wallet.Client.RPCClient
}

// AccAddress returns the account address.
func (fc *FailoverClient) AccAddress() string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.wallet.AccAddress()
}

// CurrentNodeIndex returns the index of the currently connected node.
func (fc *FailoverClient) CurrentNodeIndex() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.currentIndex
}

// CurrentRPCAddr returns the RPC address of the currently connected node.
func (fc *FailoverClient) CurrentRPCAddr() string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.nodeCfg.RPCAddrs[fc.currentIndex]
}

// CurrentGRPCAddr returns the GRPC address of the currently connected node.
func (fc *FailoverClient) CurrentGRPCAddr() string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.nodeCfg.GRPCAddrs[fc.currentIndex]
}

// FailoverCount returns the total number of failovers that have occurred.
func (fc *FailoverClient) FailoverCount() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.failoverCount
}

// NodeCount returns the total number of configured nodes.
func (fc *FailoverClient) NodeCount() int {
	return len(fc.nodeCfg.RPCAddrs)
}

// IsConnectionError checks if an error indicates a connection problem that
// should trigger a failover.
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Common connection error patterns
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"network is unreachable",
		"i/o timeout",
		"context deadline exceeded",
		"eof",
		"connection closed",
		"transport is closing",
		"server misbehaving",
		"unavailable",
		"failed to connect",
	}

	errStrLower := strings.ToLower(errStr)
	for _, pattern := range connectionErrors {
		if strings.Contains(errStrLower, pattern) {
			return true
		}
	}

	return false
}

// ExecuteWithFailover executes a function and automatically fails over
// to the next node if a connection error is detected.
func (fc *FailoverClient) ExecuteWithFailover(fn func() error) error {
	err := fn()
	if err != nil && IsConnectionError(err) {
		log.Warn().Err(err).Msg("Connection error detected, attempting failover")
		if fc.Failover() {
			// Retry after failover
			return fn()
		}
	}
	return err
}

// QueryWithFailover executes a query function and automatically fails over
// to the next node if a connection error is detected. It returns the result
// of the query function.
func QueryWithFailover[T any](fc *FailoverClient, fn func() (T, error)) (T, error) {
	result, err := fn()
	if err != nil && IsConnectionError(err) {
		log.Warn().Err(err).Msg("Connection error detected during query, attempting failover")
		if fc.Failover() {
			// Retry after failover
			return fn()
		}
	}
	return result, err
}

// HealthCheck performs a health check on the current node by querying ABCI info.
func (fc *FailoverClient) HealthCheck(ctx context.Context) error {
	fc.mu.RLock()
	rpcClient := fc.wallet.Client.RPCClient
	fc.mu.RUnlock()

	_, err := rpcClient.ABCIInfo(ctx)
	return err
}

// EnsureHealthy checks if the current node is healthy, and if not,
// attempts to failover to a healthy node.
func (fc *FailoverClient) EnsureHealthy(ctx context.Context) error {
	err := fc.HealthCheck(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Current node unhealthy, attempting failover")
		if !fc.Failover() {
			return err
		}
	}
	return nil
}
