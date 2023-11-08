package file_system

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
	"github.com/wealdtech/go-merkletree/v2"
	"github.com/wealdtech/go-merkletree/v2/sha3"
)
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func buildTree(buf io.Reader, chunkSize int64) ([]byte, []byte, [][]byte, int, error) {
	size := 0

	data := make([][]byte, 0)
	chunks := make([][]byte, 0)

	index := 0

	for {
		b := make([]byte, chunkSize)
		read, _ := buf.Read(b)

		if read == 0 {
			break
		}

		b = b[:read]

		size += read

		chunks = append(chunks, b)

		hexedData := hex.EncodeToString(b)

		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%d%s", index, hexedData))) // appending the index and the data
		hashName := hash.Sum(nil)

		data = append(data, hashName)

		index++
	}

	tree, err := merkletree.NewTree(
		merkletree.WithData(data),
		merkletree.WithHashType(sha3.New512()),
		merkletree.WithSalt(false),
	)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	r := tree.Root()

	exportedTree, err := json.Marshal(tree)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	return r, exportedTree, chunks, size, nil
}

func WriteFile(db *badger.DB, reader io.Reader, merkle []byte, owner string, start int64, address string, chunkSize int64) (size int, err error) {
	log.Info().Msg(fmt.Sprintf("Writing %x to disk", merkle))
	root, exportedTree, chunks, s, err := buildTree(reader, chunkSize)
	if err != nil {
		log.Error().Err(fmt.Errorf("cannot build tree | %w", err))
		return 0, err
	}
	size = s
	if hex.EncodeToString(merkle) != hex.EncodeToString(root) {
		return 0, fmt.Errorf("merkle does not match %x != %x", merkle, root)
	}

	err = db.Update(func(txn *badger.Txn) error {
		err = txn.Set(treeKey(merkle, owner, start), exportedTree)
		if err != nil {
			e := fmt.Errorf("cannot set tree %x | %w", merkle, err)
			log.Error().Err(e)
			return e
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	i := 0
	for len(chunks) > 0 {
		chunk := chunks[0]
		chunks = chunks[1:]

		err = db.Update(func(txn *badger.Txn) error {
			err := txn.Set(chunkKey(merkle, owner, start, i), chunk)
			if err != nil {
				e := fmt.Errorf("cannot set chunk %d | %w", i, err)
				log.Error().Err(e)
				return e
			}
			return nil
		})
		if err != nil {
			return 0, err
		}

		i++
	}

	return size, nil
}

func DeleteFile(db *badger.DB, merkle []byte, owner string, start int64) error {
	log.Info().Msg(fmt.Sprintf("Removing %x from disk...", merkle))
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(treeKey(merkle, owner, start))
		if err != nil {
			return err
		}
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := majorChunkKey(merkle, owner, start)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := txn.Delete(k)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func ListFiles(db *badger.DB) ([][]byte, []string, []int64, error) {
	merkles := make([][]byte, 0)
	owners := make([]string, 0)
	starts := make([]int64, 0)

	err := db.View(func(txn *badger.Txn) error {
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

func Dump(db *badger.DB) (map[string]string, error) {
	files := make(map[string]string)

	err := db.View(func(txn *badger.Txn) error {
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

func GetFileTreeByChunk(db *badger.DB, merkle []byte, owner string, start int64, chunkToLoad int) (newTree *merkletree.MerkleTree, chunkOut []byte, err error) {
	tree := treeKey(merkle, owner, start)
	chunk := chunkKey(merkle, owner, start, chunkToLoad)

	err = db.View(func(txn *badger.Txn) error {
		t, err := txn.Get(tree)
		if err != nil {
			return err
		}
		err = t.Value(func(val []byte) error {
			err := json.Unmarshal(val, newTree)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		c, err := txn.Get(chunk)
		if err != nil {
			return err
		}

		err = c.Value(func(val []byte) error {
			chunkOut = val
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func GetCIDFromFID(txn *badger.Txn, fid string) (cid string, err error) {
	found := false

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		if found {
			break
		}

		item := it.Item()

		_ = item.Value(func(val []byte) error {
			if string(val) == fid {
				cid = string(item.Key()[len("files/"):])

				found = true
			}

			return nil
		})

	}

	if !found {
		err = fmt.Errorf("no fid found")
	}

	return
}

func GetFileData(db *badger.DB, merkle []byte, owner string, start int64) ([]byte, error) {
	fileData := make([]byte, 0)

	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := majorChunkKey(merkle, owner, start)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				fileData = append(fileData, val...)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return fileData, err
}

func GetFileDataByMerkle(db *badger.DB, merkle []byte) ([]byte, error) {
	fileData := make([]byte, 0)
	o := ""
	var s int64
	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := majorChunkMerkleKey(merkle)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			k := item.Key()[len("chunks/"):]

			_, owner, start, err := SplitMerkle(k)
			if err != nil {
				return err
			}

			if len(o) == 0 {
				o = owner
			} else {
				if owner != o {
					return nil
				}
			}
			if s == 0 {
				s = start
			} else {
				if s != start {
					return nil
				}
			}

			err = item.Value(func(val []byte) error {
				fileData = append(fileData, val...)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return fileData, err
}
