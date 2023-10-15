package types

type UploadResponse struct {
	CID string `json:"cid"`
	FID string `json:"fid"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
