package types

import (
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type UploadResponse struct {
	Merkle []byte `json:"merkle"`
	Owner  string `json:"owner"`
	Start  int64  `json:"start"`
	CID    string `json:"cid"`
}

type UploadResponseV2 struct {
	Merkle   []byte  `json:"merkle"`
	Owner    string  `json:"owner"`
	Start    int64   `json:"start"`
	CID      string  `json:"cid"`
	Progress float64 `json:"progress"`
}

type AcceptedUploadResponse struct {
	JobID string `json:"job_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type VersionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"build"`
	ChainID string `json:"chain-id"`
}

type SpaceResponse struct {
	Total int64 `json:"total_space"`
	Used  int64 `json:"used_space"`
	Free  int64 `json:"free_space"`
}

type NetworkResponse struct {
	GRPCStatus string                  `json:"grpc-status"`
	RPCStatus  *coretypes.ResultStatus `json:"rpc-status"`
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

type PeersResponse struct {
	Peers []string `json:"peers"`
}

type HostResponse struct {
	Hosts []string `json:"hosts"`
}

type CidResponse struct {
	Cids []string `json:"cids"`
}

type CidFolderResponse struct {
	Cid  string `json:"cid"`
	Data []byte `json:"data"`
}

type CidMapResponse struct {
	CidMap map[string]string `json:"cid_map"`
}
