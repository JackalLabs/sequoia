package core

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/JackalLabs/sequoia/config"
)

const jprovIPFSStorageDir = "ipfs-storage"

func (a *App) Migrate(jprovRoot string) {
	defer log.Info().Msg("migration finished")

	cfg, err := config.Init(a.home)
	if err != nil {
		panic(err)
	}

	err = os.Rename(filepath.Join(jprovRoot, jprovIPFSStorageDir), cfg.DataDirectory)
	if err != nil {
		panic(err)
	}

	return
}
