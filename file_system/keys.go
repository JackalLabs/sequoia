package file_system

import "fmt"

func chunkKey(cid string, index int) []byte {
	return []byte(fmt.Sprintf("chunks/%s/%010d", cid, index))
}

func majorChunkKey(cid string) []byte {
	return []byte(fmt.Sprintf("chunks/%s/", cid))
}

func treeKey(cid string) []byte {
	return []byte(fmt.Sprintf("tree/%s", cid))
}

func fileKey(cid string) []byte {
	return []byte(fmt.Sprintf("files/%s", cid))
}
