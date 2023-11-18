package file_system

import "github.com/dgraph-io/badger/v4"

type FileSystem struct {
	db *badger.DB
}

func NewFileSystem(db *badger.DB) *FileSystem {
	return &FileSystem{db: db}
}

func (f *FileSystem) Close() {
	f.db.Close()
}
