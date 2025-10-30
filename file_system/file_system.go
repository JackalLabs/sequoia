package file_system

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"

	"github.com/wealdtech/go-merkletree/v2/sha3"

	sequoiaTypes "github.com/JackalLabs/sequoia/types"
	treeblake3 "github.com/wealdtech/go-merkletree/v2/blake3"
	"github.com/zeebo/blake3"

	"github.com/JackalLabs/sequoia/api/types"

	ipfslite "github.com/hsanjuan/ipfs-lite"

	"github.com/ipfs/boxo/ipld/merkledag"

	"github.com/ipfs/boxo/ipld/unixfs"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-cid"
	ipldFormat "github.com/ipfs/go-ipld-format"
	"github.com/wealdtech/go-merkletree/v2"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func BuildTree(buf io.Reader, chunkSize int64, proofType int64) ([]byte, []byte, int, error) {
	size := 0

	data := make([][]byte, 0)

	index := 0

	for {
		b := make([]byte, chunkSize)
		read, _ := buf.Read(b)

		if read == 0 {
			break
		}

		b = b[:read]

		size += read

		var h hash.Hash
		switch proofType {
		case sequoiaTypes.ProofTypeBlake3:
			h = blake3.New()
		default:
			h = sha256.New()
		}

		_, err := fmt.Fprintf(h, "%d%x", index, b) // appending the index and the data
		if err != nil {
			log.Warn().Msg("failed to write to hash")
			break
		}
		hashName := h.Sum(nil)

		data = append(data, hashName)

		index++
	}

	var h merkletree.HashType
	switch proofType {
	case sequoiaTypes.ProofTypeBlake3:
		h = treeblake3.New256()
	default:
		h = sha3.New512()
	}

	tree, err := merkletree.NewTree(
		merkletree.WithData(data),
		merkletree.WithHashType(h),
		merkletree.WithSalt(false),
	)
	if err != nil {
		return nil, nil, 0, err
	}

	r := tree.Root()

	exportedTree, err := json.Marshal(tree)
	if err != nil {
		return nil, nil, 0, err
	}

	return r, exportedTree, size, nil
}

func (f *FileSystem) WriteFile(reader sequoiaTypes.FileReader, merkle []byte, owner string, start int64, chunkSize int64, proofType int64, ipfsParams *ipfslite.AddParams) (size int, cid string, err error) {
	log.Info().Msg(fmt.Sprintf("Writing %x to disk", merkle))
	root, exportedTree, s, err := BuildTree(reader, chunkSize, proofType)
	if err != nil {
		log.Error().Err(fmt.Errorf("cannot build tree | %w", err))
		return 0, "", err
	}
	size = s
	if hex.EncodeToString(merkle) != hex.EncodeToString(root) {
		return 0, "", fmt.Errorf("merkle does not match %x != %x", merkle, root)
	}

	// Seek back to the beginning of the file so it can be read again
	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return 0, "", fmt.Errorf("cannot seek to beginning of file | %w", err)
	}

	var n ipldFormat.Node
	if proofType == sequoiaTypes.ProofTypeIPFSFolder {
		folderNode := unixfs.EmptyDirNode()
		// Read entire reader into a single byte slice called data
		data, err := io.ReadAll(reader)
		if err != nil {
			return 0, "", err
		}
		err = folderNode.UnmarshalJSON(data)
		if err != nil {
			return 0, "", err
		}

		err = f.ipfs.Add(context.Background(), folderNode)
		if err != nil {
			return 0, "", err
		}
		n = folderNode
	} else {
		n, err = f.ipfs.AddFile(context.Background(), reader, ipfsParams)
		if err != nil {
			return 0, "", err
		}
	}

	err = f.db.Update(func(txn *badger.Txn) error {
		err = txn.Set(treeKey(merkle, owner, start), exportedTree)
		if err != nil {
			e := fmt.Errorf("cannot set tree %x | %w", merkle, err)
			log.Error().Err(e)
			return e
		}

		err = txn.Set(fmt.Appendf(nil, "cid/%x", merkle), []byte(n.Cid().String()))
		if err != nil {
			e := fmt.Errorf("cannot set cid %x | %w", merkle, err)
			log.Error().Err(e)
			return e
		}

		return nil
	})
	if err != nil {
		return 0, "", err
	}

	fileCount.Inc()
	return size, n.Cid().String(), nil
}

func (f *FileSystem) WriteFileWithProgress(reader sequoiaTypes.FileReader, merkle []byte, owner string, start int64, chunkSize int64, proofType int64, ipfsParams *ipfslite.AddParams, tracker *types.UploadResponseV2) (size int, cid string, err error) {
	log.Info().Msg(fmt.Sprintf("Writing %x to disk", merkle))
	root, exportedTree, s, err := BuildTree(reader, chunkSize, proofType)
	if err != nil {
		log.Error().Err(fmt.Errorf("cannot build tree | %w", err))
		return 0, "", err
	}
	size = s
	if hex.EncodeToString(merkle) != hex.EncodeToString(root) {
		return 0, "", fmt.Errorf("merkle does not match %x != %x", merkle, root)
	}
	tracker.Progress = 60

	// Seek back to the beginning of the file so it can be read again
	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return 0, "", fmt.Errorf("cannot seek to beginning of file | %w", err)
	}

	tracker.Progress = 70
	var n ipldFormat.Node
	if proofType == sequoiaTypes.ProofTypeIPFSFolder {
		data, err := io.ReadAll(reader)
		if err != nil {
			return 0, "", err
		}

		folderNode := unixfs.EmptyDirNode()
		err = folderNode.UnmarshalJSON(data)
		if err != nil {
			return 0, "", err
		}

		err = f.ipfs.Add(context.Background(), folderNode)
		if err != nil {
			return 0, "", err
		}
		n = folderNode
	} else {
		n, err = f.ipfs.AddFile(context.Background(), reader, ipfsParams)
		if err != nil {
			return 0, "", err
		}
	}
	tracker.Progress = 90
	err = f.db.Update(func(txn *badger.Txn) error {
		err = txn.Set(treeKey(merkle, owner, start), exportedTree)
		if err != nil {
			e := fmt.Errorf("cannot set tree %x | %w", merkle, err)
			log.Error().Err(e)
			return e
		}

		err = txn.Set(fmt.Appendf(nil, "cid/%x", merkle), []byte(n.Cid().String()))

		return nil
	})
	if err != nil {
		return 0, "", err
	}

	fileCount.Inc()
	tracker.Progress = 100
	tracker.Status = "done!"

	return size, n.Cid().String(), nil
}

func (f *FileSystem) CreateIPFSFolder(childCIDs map[string]cid.Cid) (node ipldFormat.Node, err error) {
	n, err := f.GenIPFSFolderData(childCIDs)
	if err != nil {
		return nil, err
	}

	// Add the folder node to the DAG service
	err = f.ipfs.Add(context.Background(), n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (f *FileSystem) GenIPFSFolderData(childCIDs map[string]cid.Cid) (node ipldFormat.Node, err error) {
	folderNode := unixfs.EmptyDirNode()

	for key, childCID := range childCIDs {
		// Create a link
		link := &ipldFormat.Link{
			Name: key,
			Cid:  childCID,
		}

		// Add the link to the folder node
		err := folderNode.AddRawLink(key, link)
		if err != nil {
			return nil, err
		}
	}

	return folderNode, nil
}

func (f *FileSystem) DeleteFile(merkle []byte, owner string, start int64) error {
	if err := f.removeContract(merkle, owner, start); err != nil {
		return err
	}

	return nil
}

func (f *FileSystem) removeContract(merkle []byte, owner string, start int64) error {
	err := f.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(treeKey(merkle, owner, start))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Warn().Err(err)
	}

	found := false
	// check for other contracts with same file
	err = f.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: merkle})
		defer it.Close()
		it.Rewind()

		found = it.Valid()
		return nil
	})
	if err != nil {
		log.Warn().Err(err)
	}

	if !found {
		log.Debug().Hex("merkle", merkle).Msg("zero contracts tied to the file")
		return f.deleteFile(merkle)
	}

	return nil
}

func (f *FileSystem) deleteFile(merkle []byte) error {
	ctx := context.Background()
	log.Info().Msg(fmt.Sprintf("Removing %x from disk...", merkle))
	fileCount.Dec()

	// find ipfs cid
	fcid := ""
	err := f.db.View(func(txn *badger.Txn) error {
		b, err := txn.Get(fmt.Appendf(nil, "cid/%x", merkle))
		if err != nil {
			return err
		}

		return b.Value(func(val []byte) error {
			fcid = string(val)
			return nil
		})
	})
	if err != nil {
		log.Warn().Err(err)
	}

	err = f.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(fmt.Appendf(nil, "cid/%x", merkle))
	})
	if err != nil {
		log.Warn().Err(err)
	}

	c, err := cid.Decode(fcid)
	if err != nil {
		return err
	}

	return removeDAGNode(ctx, f.ipfs, c)
}

func collectDAGNodes(ctx context.Context, dagService *ipfslite.Peer, rootCid cid.Cid) (map[cid.Cid]struct{}, error) {
	nodesToVisit := []cid.Cid{rootCid}
	visitedNodes := make(map[cid.Cid]struct{})

	for len(nodesToVisit) > 0 {
		// Pop the next CID to process
		currentCid := nodesToVisit[0]
		nodesToVisit = nodesToVisit[1:]

		// Skip if already visited
		if _, alreadyVisited := visitedNodes[currentCid]; alreadyVisited {
			continue
		}

		// Mark as visited
		visitedNodes[currentCid] = struct{}{}

		// Get the node
		node, err := dagService.Get(ctx, currentCid)
		if err != nil {
			return nil, err
		}

		// Add all links to the queue
		for _, link := range node.Links() {
			if _, alreadyVisited := visitedNodes[link.Cid]; !alreadyVisited {
				nodesToVisit = append(nodesToVisit, link.Cid)
			}
		}
	}

	return visitedNodes, nil
}

func removeDAGNode(ctx context.Context, dagService *ipfslite.Peer, c cid.Cid) error {
	// Collect all nodes in the DAG using the non-recursive function
	nodesToRemove, err := collectDAGNodes(ctx, dagService, c)
	if err != nil {
		return err
	}

	bstore := dagService.BlockStore()

	for cidToRemove := range nodesToRemove {
		if err := bstore.DeleteBlock(ctx, cidToRemove); err != nil {
			return err
		}
	}

	return nil
}

func (f *FileSystem) ListFiles() ([][]byte, []string, []int64, error) {
	merkles := make([][]byte, 0)
	owners := make([]string, 0)
	starts := make([]int64, 0)

	err := f.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("tree/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				newValue := k[len(prefix):]
				merkle, owner, start, err := SplitMerkle(newValue)
				if err != nil {
					return err
				}

				merkles = append(merkles, merkle)
				owners = append(owners, owner)
				starts = append(starts, start)

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return merkles, owners, starts, err
}

func (f *FileSystem) ProcessFiles(fn func(merkle []byte, owner string, start int64)) error {
	err := f.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("tree/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				newValue := k[len(prefix):]
				merkle, owner, start, err := SplitMerkle(newValue)
				if err != nil {
					return err
				}

				fn(merkle, owner, start)

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (f *FileSystem) Dump() (map[string]string, error) {
	files := make(map[string]string)

	err := f.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()

			if string(k)[:4] == "tree" {
				continue
			}
			if string(k)[:5] == "chunk" {
				continue
			}

			err := item.Value(func(v []byte) error {
				files[string(k)] = string(v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return files, err
}

func (f *FileSystem) GetFileTreeByChunk(merkle []byte, owner string, start int64, chunkToLoad int, chunkSize int, proofType int64) (*merkletree.MerkleTree, []byte, error) {
	tree := treeKey(merkle, owner, start)

	var newTree merkletree.MerkleTree

	err := f.db.View(func(txn *badger.Txn) error {
		t, err := txn.Get(tree)
		if err != nil {
			return fmt.Errorf("cannot find tree structure | %w", err)
		}
		err = t.Value(func(val []byte) error {
			err := json.Unmarshal(val, &newTree)
			if err != nil {
				return fmt.Errorf("can't unmarshal tree | %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get tree | %w", err)
	}

	fcid := ""
	err = f.db.View(func(txn *badger.Txn) error {
		b, err := txn.Get(fmt.Appendf(nil, "cid/%x", merkle))
		if err != nil {
			return err
		}

		_ = b.Value(func(val []byte) error {
			fcid = string(val)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to match cid to merkle %w", err)
	}

	c, err := cid.Decode(fcid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode cid: %s | %w", fcid, err)
	}

	var chunkOut []byte
	if proofType == sequoiaTypes.ProofTypeIPFSFolder {
		n, err := f.ipfs.Get(context.Background(), c)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get node chunk for cid: %s | %w", fcid, err)
		}
		data, err := n.(*merkledag.ProtoNode).MarshalJSON()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to cast proto node: %s | %w", fcid, err)
		}
		chunkOut = data
	} else {
		chunkOut, err = f.ipfs.GetFileChunk(context.Background(), c, chunkToLoad, chunkSize)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get chunk from unixfs %w", err)
		}
	}

	if chunkOut == nil {
		return nil, nil, errors.New("chunk is nil, something is wrong")
	}

	return &newTree, chunkOut, nil
}

func (f *FileSystem) CheckTree(merkle []byte, owner string, start int64) bool {
	tree := treeKey(merkle, owner, start)

	found := true
	_ = f.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(tree)
		if err != nil {
			found = false
		}
		return nil
	})

	return found
}

func (f *FileSystem) GetFileData(merkle []byte) (io.ReadSeekCloser, error) {
	fcid := ""
	err := f.db.View(func(txn *badger.Txn) error {
		b, err := txn.Get(fmt.Appendf(nil, "cid/%x", merkle))
		if err != nil {
			return err
		}

		_ = b.Value(func(val []byte) error {
			fcid = string(val)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get cid mapping from disk: %w", err)
	}

	c, err := cid.Decode(fcid)
	if err != nil {
		return nil, fmt.Errorf("cannot decode cid '%s': %w", fcid, err)
	}

	rsc, err := f.ipfs.GetFile(context.Background(), c)
	if err != nil {
		if strings.Contains(err.Error(), "is a directory") {
			node, err := f.ipfs.Get(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("cannot get folder for cid '%s': %w", c.String(), err)
			}
			folder, ok := node.(*merkledag.ProtoNode)
			if !ok {
				return nil, fmt.Errorf("cannot parse folder from cid '%s'", c.String())
			}
			b, err := json.Marshal(folder)
			if err != nil {
				return nil, fmt.Errorf("cannot marshal folder for cid '%s': %w", c.String(), err)
			}
			return sequoiaTypes.NewBytesSeeker(b), nil
		}

		return nil, fmt.Errorf("cannot get file for cid '%s': %w", c.String(), err)
	}

	return rsc, err
}
