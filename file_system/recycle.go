package file_system

import (
	"bytes"
	"context"
	"fmt"
	"io"

	badger "github.com/dgraph-io/badger/v4"
)

func (f *FileSystem) salvageFile(r io.Reader, chunkSize int64) ([]byte, int, error) {
	root, _, chunks, size, err := BuildTree(r, chunkSize)
	if err != nil {
		return nil, 0, err
	}

	// TODO: there must be a better way to do this
	data := make([]byte, size)
	for _, chunk := range chunks {
		data = append(data, chunk...)
	}

	buf := bytes.NewBuffer(data)
	node, err := f.ipfs.AddFile(context.Background(), buf, nil)
	if err != nil {
		return nil, 0, err
	}

	err = f.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(fmt.Sprintf("cid/%x", root)), []byte(node.Cid().String()))
	})
	if err != nil {
		return nil, 0, err
	}

	return root, size, nil
}

func (f *FileSystem) SalvageFile(r io.Reader, chunkSize int64) (merkle []byte, size int, err error) {
	return f.salvageFile(r, chunkSize)
}
