package types

type UploadResponse struct {
	CID string `json:"cid"`
	FID string `json:"fid"`
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
