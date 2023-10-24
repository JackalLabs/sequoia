package types

type UploadResponse struct {
	Merkle []byte `json:"merkle"`
	Owner  string `json:"owner"`
	Start  int64  `json:"start"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type VersionResponse struct {
	Version string `json:"version"`
	ChainID string `json:"chain-id"`
}

type IndexResponse struct {
	Status  string `json:"status"`
	Address string `json:"address"`
}

type ListResponse struct {
	Files []string `json:"files"`
	Count int      `json:"count"`
}

type LegacyAPIListValue struct {
	CID string `json:"cid"`
	FID string `json:"fid"`
}

type LegacyAPIListResponse struct {
	Data []LegacyAPIListValue `json:"data"`
}
