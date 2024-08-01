package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JackalLabs/jackal-provider/jprov/archive"
	"github.com/JackalLabs/sequoia/config"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"

	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/v2/sha3"

	merkletree2 "github.com/wealdtech/go-merkletree/v2"
)

const maxQueryAttempt = 30

func (a *App) salvage(jprovdHome string, chunkSize int64) error {
	log.Info().Msg("salvaging started...")
	jprovArchive := archive.NewSingleCellArchive(jprovdHome)
	w, err := config.InitWallet(a.home)
	if err != nil {
		return err
	}

	cl := types.NewQueryClient(w.Client.GRPCConn)

	dirList, err := os.ReadDir(filepath.Join(jprovdHome, "storage"))
	if err != nil {
		return err
	}

	for _, d := range dirList {
		if !d.IsDir() {
			continue
		}

		log.Info().Str("fid", d.Name()).Msg("attempting to salvage file")
		err = a.salvageFile(jprovArchive, cl, chunkSize, d.Name())
		if err != nil {
			log.Error().Err(err).Str("fid", d.Name()).Msg("failed salvage file")
		}
	}

	return nil
}

func (a *App) queryAllFilesByMerkle(cl types.QueryClient, merkle []byte, attempt int8) ([]types.UnifiedFile, error) {
	if attempt > maxQueryAttempt {
		return nil, errors.New("max query attempt reached")
	}
	req := &types.QueryAllFilesByMerkle{
		Pagination: &query.PageRequest{CountTotal: true},
		Merkle:     merkle,
	}

	resp, err := cl.AllFilesByMerkle(context.Background(), req)
	if err != nil {
		attempt++
		time.Sleep(time.Second * 2)
		return a.queryAllFilesByMerkle(cl, merkle, attempt)
	}

	return resp.Files, nil
}

func (a *App) salvageFile(jprovArchive archive.Archive, cl types.QueryClient, chunkSize int64, fid string) error {
	tree, err := jprovArchive.RetrieveTree(fid)
	if err != nil {
		return err
	}

	e, err := tree.Export()
	if err != nil {
		return fmt.Errorf("failed to temp export tree | %w", err)
	}
	var me merkletree.Export
	err = json.Unmarshal(e, &me)
	if err != nil {
		return fmt.Errorf("failed to temp re-import tree | %w", err)
	}

	merkle, err := merkletree2.NewTree(
		merkletree2.WithData(me.Data),
		merkletree2.WithHashType(sha3.New512()),
		merkletree2.WithSalt(me.Salt),
	)
	if err != nil {
		return err
	}

	file, err := jprovArchive.RetrieveFile(fid)
	if err != nil {
		return err
	}

	defer file.Close()

	var attempt int8 = 0
	ufs := make([]types.UnifiedFile, 0)
	res, err := a.queryAllFilesByMerkle(cl, merkle.Root(), attempt)
	if err != nil {
		return err
	}
	ufs = append(ufs, res...)

	for _, f := range ufs {
		owner := f.Owner
		start := f.Start

		_, _, err = a.fileSystem.WriteFile(file, merkle.Root(), owner, start, "", chunkSize)
		if err != nil {
			return err
		}
	}
	return nil
}
