package ipfs

import (
	"context"
	"fmt"
	"strings"

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

func MakeIPFS(ctx context.Context, db *badger.DB, ds datastore.Batching, bs blockstore.Blockstore, port int, customDomain string) (*ipfslite.Peer, host.Host, error) {
	var key crypto.PrivKey
	_ = db.View(func(txn *badger.Txn) error {
		k, err := txn.Get(PrivateKeyKey)
		if err != nil {
			return err
		}
		_ = k.Value(func(val []byte) error {
			kk, err := crypto.UnmarshalPrivateKey(val)
			if err != nil {
				return err
			}

			key = kk
			return nil
		})
		return nil
	})

	if key == nil {
		priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
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

		key = priv
	}

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
