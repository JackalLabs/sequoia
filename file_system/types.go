package file_system

import (
	"context"

	ipfs2 "github.com/JackalLabs/sequoia/ipfs"
	"github.com/dgraph-io/badger/v4"
	ipfslite "github.com/hsanjuan/ipfs-lite"
)

type FileSystem struct {
	db   *badger.DB
	ipfs *ipfslite.Peer
}

func NewFileSystem(ctx context.Context, db *badger.DB, ipfsPort int) *FileSystem {
	ipfs, err := ipfs2.MakeIPFS(ctx, db, ipfsPort)
	if err != nil {
		panic(err)
	}
	return &FileSystem{db: db, ipfs: ipfs}
}

func (f *FileSystem) Close() {
	f.db.Close()
}
