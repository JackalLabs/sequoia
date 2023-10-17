package file_system

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/JackalLabs/sequoia/utils"
	"github.com/dgraph-io/badger/v4"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/sha3"
	"io"
)

func WriteFile(db *badger.DB, reader io.Reader, signee string, address string, cidOverride string) (merkle string, fid string, cid string, size int, err error) {
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

	if cidOverride != "" {
		cid = cidOverride
	}

	err = db.Update(func(txn *badger.Txn) error {

		chunkSize := 1024

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

		err = txn.Set(treeKey(cid), exportedTree)
		if err != nil {
			fmt.Printf("Cannot set tree %s | %e\n", cid, err)
		}

		for i, chunk := range chunks {
			err := txn.Set(chunkKey(cid, i), chunk)
			if err != nil {
				fmt.Printf("Cannot set chunk %d | %e", i, err)
			}
		}

		err = txn.Set(fileKey(cid), []byte(fid))
		if err != nil {
			fmt.Printf("Cannot set cid %s | %e\n", cid, err)
		}

		merkle = r

		return nil

	})

	return
}

func DeleteFile(db *badger.DB, cid string) error {
	fmt.Printf("Removing %s from disk...\n", cid)
	return db.Update(func(txn *badger.Txn) error {

		err := txn.Delete(treeKey(cid))
		if err != nil {
			return err
		}
		err = txn.Delete(fileKey(cid))
		if err != nil {
			return err
		}
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := majorChunkKey(cid)
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

func GetFileChunk(db *badger.DB, cid string, chunkToLoad int) (newTree *merkletree.MerkleTree, chunkOut []byte, err error) {
	tree := treeKey(cid)
	chunk := chunkKey(cid, chunkToLoad)

	err = db.View(func(txn *badger.Txn) error {
		t, err := txn.Get(tree)
		if err != nil {
			return err
		}
		err = t.Value(func(val []byte) error {
			newTree, err = merkletree.ImportMerkleTree(val, sha3.New512())
			if err != nil {
				return err
			}
			return nil
		})

		c, err := txn.Get(chunk)
		if err != nil {
			return err
		}

		err = c.Value(func(val []byte) error {
			chunkOut = val
			return nil
		})

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

func GetFileDataByFID(db *badger.DB, fid string) (file []byte, err error) {
	err = db.View(func(txn *badger.Txn) error {
		cid, err := GetCIDFromFID(txn, fid)
		if err != nil {
			return err
		}

		file, err = GetFileData(db, cid)
		if err != nil {
			return err
		}
		return nil
	})

	return

}

func GetFileData(db *badger.DB, cid string) ([]byte, error) {

	fileData := make([]byte, 0)

	err := db.View(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := majorChunkKey(cid)
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
