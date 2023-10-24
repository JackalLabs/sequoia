package file_system

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

func chunkKey(merkle []byte, owner string, start int64, index int) []byte {
	return []byte(fmt.Sprintf("chunks/%x/%s/%d/%010d", merkle, owner, start, index))
}

func majorChunkKey(merkle []byte, owner string, start int64) []byte {
	return []byte(fmt.Sprintf("chunks/%x/%s/%d/", merkle, owner, start))
}

func majorChunkMerkleKey(merkle []byte) []byte {
	return []byte(fmt.Sprintf("chunks/%x", merkle))
}

func SplitMerkle(key []byte) (merkle []byte, owner string, start int64, err error) {
	its := strings.Split(string(key), "/")
	merkle, err = hex.DecodeString(its[0])
	if err != nil {
		return
	}

	start, err = strconv.ParseInt(its[2], 10, 64)
	if err != nil {
		return
	}

	owner = its[1]
	return
}

func treeKey(merkle []byte, owner string, start int64) []byte {
	return []byte(fmt.Sprintf("tree/%x/%s/%d", merkle, owner, start))
}
