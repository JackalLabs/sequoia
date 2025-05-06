package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JackalLabs/sequoia/api"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/ipfs"
	"github.com/JackalLabs/sequoia/logger"
	"github.com/JackalLabs/sequoia/types"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

const owner = "jkl15w9zm873n0femu8egv7hyj9l7jfqtqvwyrqk73"

func writeFile(f *file_system.FileSystem, file []byte) ([]byte, uint, error) {
	root, _, _, _, err := file_system.BuildTree(bytes.NewReader(file), 1024)
	if err != nil {
		return root, 0, err
	}

	size, _, err := f.WriteFile(bytes.NewReader(file), root, owner, 0, 1024, 0, nil)
	return root, uint(size), err
}

func TestPathing(t *testing.T) {
	r := require.New(t)
	options := badger.DefaultOptions("/tmp/badger/k")
	options.Logger = &logger.SequoiaLogger{}

	db, err := badger.Open(options)
	r.NoError(err)

	err = db.DropAll()
	r.NoError(err)

	ds, err := ipfs.NewBadgerDataStore(db)
	r.NoError(err)

	f, err := file_system.NewFileSystem(context.Background(), db, "", ds, nil, 4005, "/dns4/ipfs.example.com/tcp/4001")
	r.NoError(err)

	file := []byte("happy birthday!")
	fileId, size, err := writeFile(f, file)
	r.NoError(err)

	node := types.FileNode{
		Name:   "HappyBirthday.txt",
		Merkle: fileId,
		Size:   size,
	}

	folder := types.FolderData{
		Name:    "HappyContainer",
		Version: 0,
		Children: []types.FileNode{
			node,
		},
	}

	folderData, err := json.Marshal(folder)
	r.NoError(err)

	folderId, folderSize, err := writeFile(f, folderData)
	r.NoError(err)

	folderFileNode := types.FileNode{
		Name:   folder.Name,
		Merkle: folderId,
		Size:   folderSize,
	}

	outerFolderNode := types.FolderData{
		Name:    "BiggerContainer",
		Version: 0,
		Children: []types.FileNode{
			folderFileNode,
		},
	}

	outerFolderData, err := json.Marshal(outerFolderNode)
	r.NoError(err)
	outerFolderId, _, err := writeFile(f, outerFolderData)
	r.NoError(err)

	pathData, _, err := api.GetMerklePathData(folderId, []string{"HappyBirthday.txt"}, folder.Name, f, nil, "", "/", false)
	r.NoError(err)
	r.Equal(file, pathData)

	pathData, _, err = api.GetMerklePathData(outerFolderId, []string{"HappyContainer", "HappyBirthday.txt"}, folder.Name, f, nil, "", "/", false)
	r.NoError(err)
	r.Equal(file, pathData)

	pathData, _, err = api.GetMerklePathData(outerFolderId, []string{"HappyContainer"}, folder.Name, f, nil, "", "/", false)
	r.NoError(err)
	htmlData := string(pathData)
	r.True(strings.Contains(htmlData, "</html>"))

	pathData, _, err = api.GetMerklePathData(outerFolderId, []string{"HappyContainer"}, folder.Name, f, nil, "", "/", true)
	r.NoError(err)
	htmlData = string(pathData)
	r.False(strings.Contains(htmlData, "</html>"))
}
