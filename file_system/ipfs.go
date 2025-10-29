package file_system

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-cid"
	"github.com/rs/zerolog/log"
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

// ListUnusedCids returns a list of CIDs that exist in the blockstore but are not
// referenced by any root CID in the badger database or their children.
// This recursively traverses all folder and file node children to build a complete
// set of referenced CIDs before comparing against the blockstore.
func (f *FileSystem) ListUnusedCids(ctx context.Context) ([]string, error) {
	// Step 1: Get all root CIDs from badger database
	rootCidStrings, err := f.ListCids()
	if err != nil {
		return nil, fmt.Errorf("failed to list root CIDs from badger: %w", err)
	}

	log.Info().Msgf("Found %d root CIDs in badger database", len(rootCidStrings))

	// Step 2: For each root CID, recursively collect all referenced CIDs (including children)
	referencedCids := make(map[cid.Cid]struct{})
	for _, cidStr := range rootCidStrings {
		c, err := cid.Decode(cidStr)
		if err != nil {
			log.Warn().Err(err).Str("cid", cidStr).Msg("failed to decode CID, skipping")
			continue
		}

		// Recursively collect all DAG nodes starting from this root CID
		nodes, err := collectDAGNodes(ctx, f.ipfs, c)
		if err != nil {
			log.Warn().Err(err).Str("cid", cidStr).Msg("failed to collect DAG nodes, skipping")
			continue
		}

		// Add all collected nodes to the referenced set
		for nodeCid := range nodes {
			referencedCids[nodeCid] = struct{}{}
		}
	}

	log.Info().Msgf("Found %d total referenced CIDs (including children)", len(referencedCids))

	// Step 3: Get all CIDs from the blockstore
	bstore := f.ipfs.BlockStore()
	allKeysChan, err := bstore.AllKeysChan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys from blockstore: %w", err)
	}

	blockstoreCids := make(map[cid.Cid]struct{})
	for c := range allKeysChan {
		blockstoreCids[c] = struct{}{}
	}

	log.Info().Msgf("Found %d total CIDs in blockstore", len(blockstoreCids))

	// Step 4: Find unused CIDs (in blockstore but not in referenced set)
	unusedCids := make([]string, 0)
	for blockstoreCid := range blockstoreCids {
		if _, isReferenced := referencedCids[blockstoreCid]; !isReferenced {
			unusedCids = append(unusedCids, blockstoreCid.String())
		}
	}

	log.Info().Msgf("Found %d unused CIDs", len(unusedCids))

	return unusedCids, nil
}
