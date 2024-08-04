package file_system

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (f *FileSystem) ListPeers() peer.IDSlice {
	return f.ipfsHost.Peerstore().PeersWithAddrs()
}

func (f *FileSystem) GetHosts() []string {
	peerId := f.ipfsHost.ID()

	peerString := peerId.String()

	s := make([]string, len(f.ipfsHost.Addrs()))

	for i, multiaddr := range f.ipfsHost.Addrs() {
		s[i] = fmt.Sprintf("%s/ipfs/%s", multiaddr.String(), peerString)
	}

	return s
}

func (f *FileSystem) GetCIDFromMerkle(merkle []byte) (cid string, err error) {
	err = f.db.View(func(txn *badger.Txn) error {
		c, err := txn.Get([]byte(fmt.Sprintf("cid/%x", merkle)))
		if err != nil {
			return err
		}
		err = c.Value(func(val []byte) error {
			cid = string(val)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return
}

func (f *FileSystem) ListCids() ([]string, error) {
	cids := make([]string, 0)

	err := f.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("cid/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				cids = append(cids, string(v))

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return cids, err
}

func (f *FileSystem) MapCids() (map[string][]byte, error) {
	cidMap := make(map[string][]byte)

	err := f.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("cid/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			merkle := k[len(prefix):]
			err := item.Value(func(v []byte) error {
				cidMap[string(v)] = merkle

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return cidMap, err
}
