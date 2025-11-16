package file_system

import (
	"context"
	"time"

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
	db         *badger.DB
	ipfs       *ipfslite.Peer
	ipfsHost   host.Host
	ipfsDomain string
}

func NewFileSystem(ctx context.Context, db *badger.DB, seed string, ds datastore.Batching, bs blockstore.Blockstore, ipfsPort int, ipfsDomain string) (*FileSystem, error) {
	ipfs, hh, err := ipfs2.MakeIPFS(ctx, seed, ds, bs, ipfsPort, ipfsDomain)
	if err != nil {
		return nil, err
	}
	return &FileSystem{db: db, ipfs: ipfs, ipfsHost: hh, ipfsDomain: ipfsDomain}, nil
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

// StartGC starts a background goroutine that periodically runs Badger DB garbage collection.
// This is essential to reclaim space from deleted/updated entries in the value log.
// The GC runs every 10 minutes or when the value log size exceeds 1GB.
func (f *FileSystem) StartGC() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			// Run GC with a discard ratio of 0.5 (reclaim space if 50% or more can be discarded)
			err := f.db.RunValueLogGC(0.5)
			if err != nil {
				if err == badger.ErrNoRewrite {
					// No rewrite needed, which is fine
					log.Debug().Msg("Badger GC: no rewrite needed")
				} else {
					log.Warn().Err(err).Msg("Badger GC failed")
				}
			} else {
				log.Info().Msg("Badger GC completed successfully")
			}
		}
	}()
	log.Info().Msg("Badger DB garbage collection started (runs every 10 minutes)")
}
