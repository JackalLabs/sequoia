package ipfs

import (
	"github.com/dgraph-io/badger/v4"
	ds "github.com/ipfs/go-datastore"
	bds "github.com/ipfs/go-ds-badger2"
	fds "github.com/ipfs/go-ds-flatfs"
)

func NewFlatfsDataStore(path string) (ds.Batching, error) {
	return fds.CreateOrOpen(path, fds.IPFS_DEF_SHARD, true)
}

func NewBadgerDataStore(db *badger.DB) (ds.Batching, error) {
	return bds.NewDatastoreFromDB(db)
}
