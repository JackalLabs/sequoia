package recycle

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/JackalLabs/jackal-provider/jprov/archive"
	query "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"
)

func (r *RecycleDepot) salvageFile(jprovArhcive archive.Archive, fid string) ([]byte, int, error) {
	file, err := jprovArhcive.RetrieveFile(fid)
	defer file.Close()
	if err != nil {
		return nil, 0, err
	}

	return r.fs.SalvageFile(file, r.chunkSize)
}

func (r *RecycleDepot) SalvageFiles(jprovdHome string) error {
	log.Info().Msg("salvaging jprovd files...")
	recordFile, err := os.OpenFile(
		filepath.Join(r.homeDir, salvageRecordFileName),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return err
	}

	defer recordFile.Close()

	// only used to retrieve files
	jprovArchive := archive.NewSingleCellArchive(jprovdHome)

	dirList, err := os.ReadDir(filepath.Join(jprovdHome, "storage"))
	if err != nil {
		log.Error().Err(err).Msg("failed to read jprovd storage directory")
		return err
	}

	salvaged := 0
	for _, d := range dirList {
		if !d.IsDir() {
			continue
		}

		merkle, size, err := r.salvageFile(jprovArchive, d.Name())
		if err != nil {
			log.Error().Err(err).Str("fid", d.Name()).Msg("failed to salvage file")
			continue
		}

		log.Info().
			Hex("merkle", merkle).
			Str("fid", d.Name()).
			Msg("successfully salvaged file")

		err = record(recordFile, merkle, size, d.Name())

		if err != nil {
			log.Error().
				Err(err).
				Hex("merkle", merkle).
				Str("fid", d.Name()).
				Int("size", size).
				Str("record", fmt.Sprintf("%x,%d,%s", merkle, size, d.Name())).
				Msg("failed to record salvage info")
		}

		salvaged++
	}

	log.Info().Int("count", salvaged).Msg("salvaging finished...")
	return nil
}

func record(file io.Writer, merkle []byte, size int, fid string) error {
	_, err := file.Write([]byte(fmt.Sprintf("%x,%d,%s\n", merkle, size, fid)))
	return err
}

func (r *RecycleDepot) collectOpenFiles() ([]types.UnifiedFile, error) {
	req := &types.QueryOpenFiles{
		Pagination: &query.PageRequest{CountTotal: true},
	}
	resp, err := r.queryClient.OpenFiles(context.Background(), req)
	if err != nil {
		return nil, err
	}

	uf := make([]types.UnifiedFile, resp.Pagination.GetTotal())
	uf = append(uf, resp.Files...)

	for len(resp.Pagination.GetNextKey()) != 0 {
		req.Pagination.Key = resp.Pagination.GetNextKey()
		resp, err = r.queryClient.OpenFiles(context.Background(), req)
		if err != nil {
			return uf, err
		}
		uf = append(uf, resp.Files...)
	}

	return uf, nil
}

func (r *RecycleDepot) activateFile(openFile types.UnifiedFile) (size int, cid string, err error) {
	merkle := hex.EncodeToString(openFile.Merkle)
	fileData, err := r.fs.GetFileData([]byte(merkle))
	if err != nil {
		return 0, "", err
	}

	buf := bytes.NewBuffer(fileData)
	return r.fs.WriteFile(
		buf,
		[]byte(merkle),
		openFile.Owner,
		openFile.Start,
		"",
		r.chunkSize)
}

func (r *RecycleDepot) recycleFiles() error {
	openFiles, err := r.collectOpenFiles()
	if len(openFiles) == 0 && err != nil {
		log.Error().Err(err).Msg("failed to query open files from chain")
		return err
	}

	// recycle open files that it managed to collect
	for _, openFile := range openFiles {
		size, cid, err := r.activateFile(openFile)
		if err != nil {
			log.Error().
				Err(err).
				Int("size", size).
				Str("cid", cid).
				Bytes("merkle_bytes", openFile.Merkle).
				Hex("merkle_hex", openFile.Merkle).
				Int64("start_at", openFile.Start).
				Str("owner", openFile.Owner).
				Int64("expires", openFile.Expires).
				Msg("failed to activate file")

		} else {
			log.Info().
				Int("size", size).
				Str("cid", cid).
				Bytes("merkle_bytes", openFile.Merkle).
				Hex("merkle_hex", openFile.Merkle).
				Int64("start_at", openFile.Start).
				Str("owner", openFile.Owner).
				Int64("expires", openFile.Expires).
				Msg("file successfully activated")
		}
	}

	if err != nil {
		log.Error().
			Err(err).
			Int("count", len(openFiles)).
			Msg("error occurred while attempting to collect(query) more open files from chain")
		return err
	}
	return nil
}

func (r *RecycleDepot) Start(checkInterval int64) {
	for {
		select {
		case <-r.stop:
			log.Info().Msg("shutting down recycle depot...")
			err := r.Close()
			if err != nil {
				log.Error().Err(err).Msg("error while closing recycle depot db")
			}
			return
		default:
		}

		err := r.recycleFiles()
		sleepDuration := time.Second * time.Duration(checkInterval)
		if err != nil {
			sleepDuration = sleepDuration / 2
		}
		time.Sleep(sleepDuration)
	}
}

func (r *RecycleDepot) Stop() {
	r.stop <- struct{}{}
	return
}

func (r *RecycleDepot) Close() error {
	return nil
}
