package file_system

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

func (f *FileSystem) ListPeers() []string {
	results := make([]string, 0)

	ids := f.ipfsHost.Peerstore().PeersWithAddrs()
	for _, id := range ids {
		addrs := f.ipfsHost.Peerstore().Addrs(id)
		for _, multiaddr := range addrs {
			results = append(results, fmt.Sprintf("%s/ipfs/%s", multiaddr.String(), id.String()))
		}
	}

	return results
}

func (f *FileSystem) GetHosts() []string {
	peerId := f.ipfsHost.ID()
	peerString := peerId.String()

	// Get addresses from the host
	addrs := f.ipfsHost.Addrs()
	results := make([]string, 0, len(addrs)+1) // +1 for possible custom domain

	// Add all bound addresses
	for _, multiaddr := range addrs {
		results = append(results, fmt.Sprintf("%s/ipfs/%s", multiaddr.String(), peerString))
	}

	ipfsDomain := f.ipfsDomain

	if !strings.Contains(ipfsDomain, "example.com") && len(ipfsDomain) > 2 {
		if !strings.HasPrefix(ipfsDomain, "/") {
			ipfsDomain = fmt.Sprintf("/%s", ipfsDomain)
		}

		// Add the custom domain to the results
		results = append(results, fmt.Sprintf("%s/ipfs/%s", ipfsDomain, peerString))
	}

	return results
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

	return cid, err
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
