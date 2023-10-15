package file_system

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/sha3"
	"io"
	"sequoia/utils"
)

func WriteFile(db *badger.DB, reader io.Reader, signee string, address string) (merkle string, fid string, cid string, size int, err error) {
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)
	fid, err = utils.MakeFid(tee)
	if err != nil {
		return
	}

	cid, err = utils.MakeCid(signee, address, fid)
	if err != nil {
		return
	}

	err = db.Update(func(txn *badger.Txn) error {
		chunkSize := 1024

		data := make([][]byte, 0)
		chunks := make([][]byte, 0)

		index := 0

		for {
			b := make([]byte, chunkSize)
			read, err := buf.Read(b)
			if err != nil {
				break
			}
			if read == 0 {
				break
			}
			if read < chunkSize {
				b = b[:read]
			}
			size += read

			chunks = append(chunks, b)

			hexedData := hex.EncodeToString(b)

			hash := sha256.New()
			hash.Write([]byte(fmt.Sprintf("%d%s", index, hexedData))) // appending the index and the data
			hashName := hash.Sum(nil)

			data = append(data, hashName)

			index++
		}

		tree, err := merkletree.NewUsing(data, sha3.New512(), false)
		if err != nil {
			return err
		}

		r := hex.EncodeToString(tree.Root())

		exportedTree, err := tree.Export()
		if err != nil {
			return err
		}

		tree = nil // for GC

		txn.Set(treeKey(cid), exportedTree)
		for i, chunk := range chunks {
			txn.Set(chunkKey(cid, i), chunk)
		}
		txn.Set(fileKey(cid), []byte(fid))

		merkle = r

		return nil

	})

	return
}

func ListFiles(db *badger.DB) ([]string, error) {

	files := make([]string, 0)

	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("files/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				files = append(files, string(k[len(prefix):]))
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
