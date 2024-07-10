package file_system

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"

	ipfs2 "github.com/JackalLabs/sequoia/ipfs"
	"github.com/dgraph-io/badger/v4"
	ipfslite "github.com/hsanjuan/ipfs-lite"
)

type FileSystem struct {
	db       *badger.DB
	ipfs     *ipfslite.Peer
	ipfsHost host.Host
}

func NewFileSystem(ctx context.Context, db *badger.DB, ipfsPort int, ipfsDomain string) (*FileSystem, error) {
	ipfs, host, err := ipfs2.MakeIPFS(ctx, db, ipfsPort, ipfsDomain)
	if err != nil {
		return nil, err
	}
	return &FileSystem{db: db, ipfs: ipfs, ipfsHost: host}, nil
}

func (f *FileSystem) Close() {
	f.db.Close()
}
