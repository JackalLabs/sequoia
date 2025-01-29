package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/JackalLabs/sequoia/utils"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/stretchr/testify/require"
)

func TestNotes(t *testing.T) {
	r := require.New(t)

	params := ipfslite.AddParams{
		Layout:    "balanced",
		Chunker:   "size-256",
		RawLeaves: true,
		Hidden:    false,
		Shard:     false,
		NoCopy:    false,
		HashFun:   "sha-256",
	}
	p, err := json.Marshal(params)
	r.NoError(err)

	s := make(map[string]string)
	s["ipfsParams"] = string(p)

	js, err := json.Marshal(s)
	r.NoError(err)

	f := types.UnifiedFile{
		Note: string(js),
	}

	pars := utils.GetIPFSParams(&f)
	t.Log(pars)

	r.NotNil(pars)

	r.Equal(params, *pars)
}
