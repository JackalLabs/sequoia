package types

type FileNode struct {
	Name   string `json:"name"`
	Merkle []byte `json:"merkle"`
	Size   uint   `json:"size"`
}

type FolderData struct {
	Name     string     `json:"name"`
	Merkle   []byte     `json:"merkle"`
	Version  uint       `json:"version"`
	Children []FileNode `json:"children"`
}
