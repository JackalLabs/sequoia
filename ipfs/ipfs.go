package ipfs

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"

	bds "github.com/ipfs/go-ds-badger2"
)

func MakeIPFS(ctx context.Context, db *badger.DB, port int, customDomain string) (*ipfslite.Peer, error) {
	ds, err := bds.NewDatastoreFromDB(db)
	if err != nil {
		return nil, err
	}

	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, err
	}

	listen, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	m := []multiaddr.Multiaddr{listen}

	if !strings.Contains(customDomain, "example.com") && len(customDomain) > 2 {
		domainListener, _ := multiaddr.NewMultiaddr(customDomain)
		m = []multiaddr.Multiaddr{listen, domainListener}
	}

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		m,
		ds,
		ipfslite.Libp2pOptionsExtra...,
	)
	if err != nil {
		return nil, err
	}

	lite, err := ipfslite.New(ctx, ds, nil, h, dht, nil)
	if err != nil {
		return nil, err
	}

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	return lite, nil
}
