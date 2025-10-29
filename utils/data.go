package utils

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func OpenBadger(dataDir string) (*badger.DB, error) {
	options := badger.DefaultOptions(dataDir)
	options = options.WithBlockCacheSize(256 << 22).WithMaxLevels(8)

	badgerLogLevel := badger.INFO
	switch log.Logger.GetLevel() {
	case zerolog.DebugLevel:
		badgerLogLevel = badger.DEBUG
	case zerolog.InfoLevel:
		badgerLogLevel = badger.INFO
	case zerolog.WarnLevel:
		badgerLogLevel = badger.WARNING
	case zerolog.ErrorLevel:
		badgerLogLevel = badger.ERROR
	}
	log.Info().
		Int("badger_log_level", int(badgerLogLevel)).
		Str("global_log_level", log.Logger.GetLevel().String()).
		Msg("badger logging configured")

	options = options.WithLoggingLevel(badgerLogLevel)

	log.Info().Msg("Creating sequoia app...")

	db, err := badger.Open(options)
	if err != nil {
		log.Error().Err(err).Msg("Error opening database")
		return nil, err
	}
	log.Info().Msg("Opened database")

	return db, nil
}
