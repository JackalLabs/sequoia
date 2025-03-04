package file_system

import (
	"context"

	"github.com/ipfs/boxo/blockstore"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	ipfs2 "github.com/JackalLabs/sequoia/ipfs"
	"github.com/dgraph-io/badger/v4"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/rs/zerolog/log"
)

type FileSystem struct {
	db       *badger.DB
	ipfs     *ipfslite.Peer
	ipfsHost host.Host
}

func NewFileSystem(ctx context.Context, db *badger.DB, seed string, ds datastore.Batching, bs blockstore.Blockstore, ipfsPort int, ipfsDomain string) (*FileSystem, error) {
	ipfs, hh, err := ipfs2.MakeIPFS(ctx, db, seed, ds, bs, ipfsPort, ipfsDomain)
	if err != nil {
		return nil, err
	}
	return &FileSystem{db: db, ipfs: ipfs, ipfsHost: hh}, nil
}

func (f *FileSystem) Close() {
	err := f.db.Close()
	if err != nil {
		log.Error().Err(err).Msg("error occurred while closing database")
	}
	err = f.ipfsHost.Close()
	if err != nil {
		log.Error().Err(err).Msg("error occurred while stopping ipfs host")
	}
}

func (f *FileSystem) Connect(info *peer.AddrInfo) {
	log.Info().Msgf("Attempting connection to %s", info.String())
	err := f.ipfsHost.Connect(context.Background(), *info)
	if err != nil {
		log.Warn().Msgf("Could not connect to %s | %v", info.String(), err)
	}
}
