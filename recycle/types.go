package recycle

import (
	fs "github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/proofs"
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
	prover      *proofs.Prover
}

func NewRecycleDepot(home string, address string, chunkSize int64, fs *fs.FileSystem, prover *proofs.Prover, queryClient types.QueryClient) (*RecycleDepot, error) {
	return &RecycleDepot{
		fs:          fs,
		homeDir:     home,
		chunkSize:   chunkSize,
		stop:        false,
		queryClient: queryClient,
		address:     address,
		prover:      prover,
	}, nil
}
