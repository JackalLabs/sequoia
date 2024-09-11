package ipfs

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/boxo/blockstore"
	ds "github.com/ipfs/go-datastore"
	bds "github.com/ipfs/go-ds-badger2"
	fds "github.com/ipfs/go-ds-flatfs"
)

func NewFlatfsBlockStore(path string) (blockstore.Blockstore, error) {
	ds, err := fds.CreateOrOpen(path, fds.IPFS_DEF_SHARD, true)
	if err != nil {
		return nil, err
	}

	return blockstore.NewBlockstore(ds, blockstore.Option{}), nil
}

func NewBadgerDataStore(db *badger.DB) (ds.Batching, error) {
	return bds.NewDatastoreFromDB(db)
}
