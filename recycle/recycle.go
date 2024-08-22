package recycle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JackalLabs/jackal-provider/jprov/archive"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)

func (r *RecycleDepot) salvageFile(jprovArchive archive.Archive, fid string) ([]byte, int, error) {
	file, err := jprovArchive.RetrieveFile(fid)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	return r.fs.SalvageFile(file, r.chunkSize)
}

func (r *RecycleDepot) lastSalvagedFile(record *os.File) (string, error) {
	// read backwards from the end of the file to find the last salvaged file
	line := ""
	var cursor int64 = 0
	stat, _ := record.Stat()
	filesize := stat.Size()

	if filesize == 0 {
		return "", nil
	}

	for {
		cursor -= 1
		_, err := record.Seek(cursor, io.SeekEnd)
		if err != nil {
			return "", err
		}

		char := make([]byte, 1)
		_, err = record.Read(char)
		if err != nil {
			return "", err
		}

		if cursor != -1 && (char[0] == 10) {
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line)

		if cursor == -filesize {
			break
		}
	}

	if len(line) == 0 {
		return "", nil
	}

	substrs := strings.Split(line, ",")

	record.Seek(0, io.SeekEnd)

	fid := strings.TrimSuffix(substrs[2], "\n")

	return fid, nil
}

func countJklFiles(dirList []fs.DirEntry) int64 {
	var c int64 = 0

	for _, d := range dirList {
		if !d.IsDir() {
			continue
		}

		if strings.HasPrefix(d.Name(), "jklf") {
			c++
		}
	}

	return c
}

func (r *RecycleDepot) SalvageFiles(jprovdHome string) error {
	log.Info().Msg("salvaging jprovd files...")
	recordFile, err := os.OpenFile(
		filepath.Join(r.homeDir, salvageRecordFileName),
		os.O_RDWR|os.O_CREATE,
		0o644,
	)
	if err != nil {
		return err
	}

	defer recordFile.Close()

	// only used to retrieve files
	jprovArchive := archive.NewSingleCellArchive(jprovdHome)
	fPath := filepath.Join(jprovdHome, "storage")
	log.Info().Msgf("Reading recycled files from %s", fPath)
	dirList, err := os.ReadDir(fPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to read jprovd storage directory")
		return err
	}
	r.TotalJprovFiles = countJklFiles(dirList)

	lastSalvaged, err := r.lastSalvagedFile(recordFile)
	if err != nil {
		return err
	}
	lastSalvagedFound := false

	r.SalvagedFilesCount = 0
	for _, d := range dirList {
		if !d.IsDir() {
			continue
		}

		if lastSalvaged != "" && !lastSalvagedFound {
			if d.Name() == lastSalvaged {
				lastSalvagedFound = true
			}
			r.SalvagedFilesCount++
			continue // skip the last salvage record to avoid duplicate
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

		r.SalvagedFilesCount++
	}

	log.Info().Int("count", int(r.SalvagedFilesCount)).Msg("salvaging finished...")
	return nil
}

func record(file io.Writer, merkle []byte, size int, fid string) error {
	_, err := file.Write([]byte(fmt.Sprintf("%x,%d,%s\n", merkle, size, fid)))
	return err
}

func (r *RecycleDepot) collectOpenFiles() ([]types.UnifiedFile, error) {
	req := &types.QueryOpenFiles{
		ProviderAddress: r.address,
		Pagination:      nil,
	}
	resp, err := r.queryClient.OpenFiles(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("We found %d files open", len(resp.Files))

	return resp.Files, nil
}

func (r *RecycleDepot) activateFile(openFile types.UnifiedFile) (size int, cid string, err error) {
	fileData, err := r.fs.GetFileData(openFile.Merkle)
	if err != nil {
		return 0, "", fmt.Errorf("can not get file data | %w", err)
	}

	buf := bytes.NewBuffer(fileData)
	size, cid, err = r.fs.WriteFile(
		buf,
		openFile.Merkle,
		openFile.Owner,
		openFile.Start,
		"",
		r.chunkSize)
	if err != nil {
		return 0, "", fmt.Errorf("could not write file | %w", err)
	}

	_ = r.prover.PostProof(openFile.Merkle, openFile.Owner, openFile.Start, openFile.Start, time.Now())

	return
}

func (r *RecycleDepot) recycleFiles() error {
	log.Info().Msg("Trying to recycle files...")
	openFiles, err := r.collectOpenFiles()
	if err != nil {
		log.Error().Err(err).Msg("failed to query open files from chain")
		return err
	}
	if len(openFiles) == 0 {
		log.Error().Msg("there are no open files")
		return err
	}

	// recycle open files that it managed to collect
	for _, openFile := range openFiles {

		if openFile.ContainsProver(r.address) {
			continue
		}

		log.Info().Msgf("Trying to recycle %x...", openFile.Merkle)

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
		if r.stop {
			log.Info().Msg("shutting down recycle depot...")
			err := r.Close()
			if err != nil {
				log.Error().Err(err).Msg("error while closing recycle depot db")
			}
			return
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
	r.stop = true
}

func (r *RecycleDepot) Close() error {
	return nil
}
