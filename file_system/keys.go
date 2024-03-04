package file_system

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

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
