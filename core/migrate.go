package core

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	query "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/jackalLabs/canine-chain/v3/x/storage/types"

	badger "github.com/dgraph-io/badger/v4"

	"github.com/JackalLabs/sequoia/config"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/logger"
)

const jprovStorageDir = "storage"

const jprovFileExt = ".jkl"

var jprovBlockFileNameRegex *regexp.Regexp = regexp.MustCompile("[0-9]+.jkl")

func NewV3App(home string) *App {
	cfg, err := config.Init(home)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	dataDir := os.ExpandEnv(cfg.DataDirectory)

	err = os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	options := badger.DefaultOptions(dataDir)

	options.Logger = &logger.SequoiaLogger{}
	options.BlockCacheSize = 256 << 25

	db, err := badger.Open(options)
	if err != nil {
		panic(err)
	}

	f := file_system.NewFileSystem(ctx, db, cfg.APICfg.IPFSPort)

	return &App{
		fileSystem: f,
		home:       home,
	}
}

func (a *App) Migrate(jprovRoot string) {
	defer log.Info().Msg("migration finished")

	w, err := config.InitWallet(a.home)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize wallet")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}

	params, err := a.GetStorageParams(w.Client.GRPCConn)
	if err != nil {
		log.Error().Err(err).Msg("failed to get chunk size")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}

	storagePath := filepath.Join(jprovRoot, jprovStorageDir)
	dirFile, err := os.Open(storagePath)
	if err != nil {
		log.Error().Str("directory", storagePath).Err(err).Msg("failed to open directory")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}
	defer dirFile.Close()

	dirs, err := dirFile.Readdirnames(0) // 0: read all dir contents
	if err != nil {
		log.Error().Str("directory", storagePath).Err(err).Msg("failed to read storage directory")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}
	if len(dirs) == 0 {
		log.Info().Msg("no files to migrate")
	}

	log.Info().Str("directory", storagePath).Int("files", len(dirs)).Msg("old files found under the storage directory")

	activeDeals, err := a.getOnlyMyActiveDeals(w)
	if err != nil {
		log.Error().Err(err).Msg("failed to query active deals")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}

	err = a.migrateFiles(params.ChunkSize, storagePath, activeDeals)
	if err != nil {
		log.Error().Err(err).Msg("failed to migrate files")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}

	return
}

func (a *App) getOnlyMyActiveDeals(wallet *wallet.Wallet) ([]types.LegacyActiveDeals, error) {
	q := types.NewQueryClient(wallet.Client.GRPCConn)

	providerAddr := wallet.AccAddress()

	req := types.QueryAllActiveDealsRequest{
		Pagination: &query.PageRequest{CountTotal: true},
	}

	activeDeals := make([]types.LegacyActiveDeals, 0)

	resp, err := q.ActiveDealsAll(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	for _, a := range resp.ActiveDeals {
		if a.Provider == providerAddr {
			activeDeals = append(activeDeals, a)
		}
	}

	for len(resp.Pagination.GetNextKey()) != 0 {
		req = types.QueryAllActiveDealsRequest{
			Pagination: &query.PageRequest{Key: resp.Pagination.GetNextKey()},
		}

		r, err := q.ActiveDealsAll(context.Background(), &req)
		if err != nil {
			time.Sleep(time.Second * 60) // we wait for a full minute if the request fails and try again
			continue
		}
		resp = r // we only update the pagination key if the request was successful
		for _, a := range resp.ActiveDeals {
			if a.Provider == providerAddr {
				activeDeals = append(activeDeals, a)
			}
		}
	}

	return activeDeals, nil
}

func (a *App) migrateFiles(chunkSize int64, jprovStoragePath string, activeDeals []types.LegacyActiveDeals) error {
	if len(activeDeals) == 0 {
		return nil
	}

	for _, deal := range activeDeals {
		err := a.migrateFile(chunkSize, jprovStoragePath, deal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) migrateFile(chunkSize int64, jprovStoragePath string, activeDeal types.LegacyActiveDeals) error {
	fileName := activeDeal.Fid + jprovFileExt

	merkle, err := hex.DecodeString(activeDeal.Merkle)
	if err != nil {
		return err
	}

	start, err := strconv.ParseInt(activeDeal.Startblock, 10, 64)
	if err != nil {
		return err
	}

	file, err := os.Open(filepath.Join(jprovStoragePath, fileName))
	if err == nil {
		defer file.Close()

		_, _, err := a.fileSystem.WriteFile(
			file,
			merkle,
			activeDeal.Signee,
			start,
			activeDeal.Signee,
			chunkSize,
		)
		if err != nil {
			return err
		}

		return nil
	} else {
		log.Info().Err(err).Str("fid", activeDeal.Fid).Msg("this fid is not a signle cell file version")
	}

	last, err := findLastBlock(filepath.Join(jprovStoragePath, activeDeal.Fid))
	if err != nil {
		return err
	}

	fidDir := filepath.Join(jprovStoragePath, activeDeal.Fid)
	f, err := os.OpenFile(
		filepath.Join(fidDir, blockFileName(0)),
		os.O_APPEND|os.O_WRONLY,
		0600,
	)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 1; i < last; i++ {
		path := filepath.Join(fidDir, blockFileName(i))
		if err := combine(f, path); err != nil {
			return err
		}
	}

	_, _, err = a.fileSystem.WriteFile(
		f,
		merkle,
		activeDeal.Signee,
		start,
		activeDeal.Signee,
		chunkSize,
	)
	if err != nil {
		return err
	}

	return nil
}

// combine opens source file and copy its contents into destination
func combine(dst io.Writer, srcFileName string) error {
	src, err := os.Open(srcFileName)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, src.Close())
	}()

	_, err = io.Copy(dst, src)
	return err
}

func blockFileName(index int) string {
	var name strings.Builder
	_, _ = name.WriteString(strconv.Itoa(index)) // returns length of s and a nil err
	_, _ = name.WriteString(jprovFileExt)        // returns length of s and a nil err

	return name.String()
}

// Get all files' name in directory
// An error is returned if the directory contains more directory
func findLastBlock(dir string) (int, error) {
	dirEntry, err := os.ReadDir(dir)
	if err != nil {
		return -1, err
	}

	last := 0

	for _, d := range dirEntry {
		if d.IsDir() {
			continue
		}

		if jprovBlockFileNameRegex.Match([]byte(d.Name())) {
			last++
		}
	}

	return last, nil
}
