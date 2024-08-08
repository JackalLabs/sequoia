package recycle

import (
	fs "github.com/JackalLabs/sequoia/file_system"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
)

const salvageRecordFileName = "salvage_record"

type RecycleDepot struct {
	fs          *fs.FileSystem
	stop        bool
	chunkSize   int64
	homeDir     string
	queryClient types.QueryClient
	address     string
}

func NewRecycleDepot(home string, address string, chunkSize int64, fs *fs.FileSystem, queryClient types.QueryClient) (*RecycleDepot, error) {
	return &RecycleDepot{
		fs:          fs,
		homeDir:     home,
		chunkSize:   chunkSize,
		stop:        false,
		queryClient: queryClient,
		address:     address,
	}, nil
}
