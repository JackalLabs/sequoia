package ipfs

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/dgraph-io/badger/v4"
	crypto "github.com/libp2p/go-libp2p/core/crypto"

	"github.com/libp2p/go-libp2p"

	"github.com/libp2p/go-libp2p/core/host"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/boxo/blockstore"
	"github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multiaddr"
)

var PrivateKeyKey = []byte("IPFS_KEYS_PRIVATE")

func MakeIPFS(ctx context.Context, db *badger.DB, seed string, ds datastore.Batching, bs blockstore.Blockstore, port int, customDomain string) (*ipfslite.Peer, host.Host, error) {
	log.Info().Msg("No key was found, generating a new IPFS key...")
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, strings.NewReader(seed))
	if err != nil {
		return nil, nil, err
	}

	privOut, err := priv.Raw()
	if err != nil {
		return nil, nil, err
	}

	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set(PrivateKeyKey, privOut)
	})
	if err != nil {
		return nil, nil, err
	}
	key := priv

	defaultPort, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make ipv4 ipfs address | %w", err)
	}

	listen, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make ipv4 ipfs address | %w", err)
	}

	listen6, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip6/::/tcp/%d", port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make ipv6 ipfs address | %w", err)
	}

	m := []multiaddr.Multiaddr{listen, listen6, defaultPort}

	if !strings.Contains(customDomain, "example.com") && len(customDomain) > 2 {
		if !strings.HasPrefix(customDomain, "/") {
			customDomain = fmt.Sprintf("/%s", customDomain)
		}
		domainListener, err := multiaddr.NewMultiaddr(customDomain)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to make domain based ipfs address | %w", err)
		}
		m = append(m, domainListener)
	}

	opts := libp2p.ChainOptions(
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
	)

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		key,
		nil,
		m,
		ds,
		append(ipfslite.Libp2pOptionsExtra, opts)...,
	)
	if err != nil {
		return nil, h, err
	}

	lite, err := ipfslite.New(ctx, ds, bs, h, dht, nil)
	if err != nil {
		return nil, h, err
	}

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	return lite, h, nil
}
