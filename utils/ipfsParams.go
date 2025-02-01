package utils

import (
	"encoding/json"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)

func GetIPFSParams(file *types.UnifiedFile) *ipfslite.AddParams {
	var data map[string]any
	err := json.Unmarshal([]byte(file.Note), &data)
	if err != nil {
		log.Warn().Msgf("Could not parse note for %x", file.Merkle)
		return nil
	}

	ipfsParams, ok := data["ipfsParams"]
	if !ok {
		return nil
	}

	ip, ok := ipfsParams.(string)
	if !ok {
		return nil
	}

	log.Info().Msgf("PARAMS: %s", ip)

	var params ipfslite.AddParams
	err = json.Unmarshal([]byte(ip), &params)
	if err != nil {
		log.Warn().Msgf("Could not parse params from note for %x", file.Merkle)
		return nil
	}

	return &params
}
