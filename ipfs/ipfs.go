package ipfs

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"strings"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	bds "github.com/ipfs/go-ds-badger2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/ipfs/boxo/blockstore"
	datastore "github.com/ipfs/go-datastore"
)

func MakeIPFS(ctx context.Context, ds datastore.Batching, bs blockstore.Blockstore, port int, customDomain string) (*ipfslite.Peer, host.Host, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, nil, err
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
		priv,
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
