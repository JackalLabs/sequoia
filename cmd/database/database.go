package database

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/config"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/ipfs"
	"github.com/JackalLabs/sequoia/utils"
	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/boxo/blockstore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func DataCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "data",
		Short: "Data subcommands",
	}

	c.AddCommand(keysCmd(), getObjectCmd(), garbageCmd(), unusedCidsCmd())

	return c
}

func keysCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "keys",
		Short: "Keys subcommands",
	}

	c.AddCommand(dumpKeysCmd())

	return c
}

func dumpKeysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dump",
		Short: "Dump all keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfg, err := config.Init(home)
			if err != nil {
				return err
			}

			dataDir := os.ExpandEnv(cfg.DataDirectory)

			err = os.MkdirAll(dataDir, os.ModePerm)
			if err != nil {
				return err
			}

			db, err := utils.OpenBadger(dataDir)
			if err != nil {
				return err
			}

			return db.View(func(txn *badger.Txn) error {
				opts := badger.DefaultIteratorOptions
				opts.PrefetchValues = false // Crucial for listing keys only
				opts.PrefetchSize = 1000    // Adjust based on your workload/memory, default is 100

				it := txn.NewIterator(opts)
				defer it.Close()

				count := 0
				for it.Rewind(); it.Valid(); it.Next() {
					item := it.Item()
					key := item.Key() // This is zero-copy
					fmt.Println(string(key))
					count++
				}
				return nil
			})
		},
	}
}

func getObjectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Print object from key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfg, err := config.Init(home)
			if err != nil {
				return err
			}

			dataDir := os.ExpandEnv(cfg.DataDirectory)

			err = os.MkdirAll(dataDir, os.ModePerm)
			if err != nil {
				return err
			}

			db, err := utils.OpenBadger(dataDir)
			if err != nil {
				return err
			}

			return db.View(func(txn *badger.Txn) error {
				val, err := txn.Get([]byte(key))
				if err != nil {
					return err
				}

				return val.Value(func(val []byte) error {
					fmt.Println(string(val))
					return nil
				})
			})
		},
	}
}

func garbageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flatten",
		Short: "Flatten database",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfg, err := config.Init(home)
			if err != nil {
				return err
			}

			dataDir := os.ExpandEnv(cfg.DataDirectory)

			err = os.MkdirAll(dataDir, os.ModePerm)
			if err != nil {
				return err
			}

			db, err := utils.OpenBadger(dataDir)
			if err != nil {
				return err
			}

			return db.Flatten(runtime.NumCPU())
		},
	}
}

func unusedCidsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unused-cids",
		Short: "List unused cids",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfg, err := config.Init(home)
			if err != nil {
				return err
			}

			ctx := context.Background()

			dataDir := os.ExpandEnv(cfg.DataDirectory)

			err = os.MkdirAll(dataDir, os.ModePerm)
			if err != nil {
				return err
			}

			db, err := utils.OpenBadger(dataDir)
			if err != nil {
				return err
			}

			ds, err := ipfs.NewBadgerDataStore(db)
			if err != nil {
				return err
			}
			log.Info().Msg("Data store initialized")

			bsDir := os.ExpandEnv(cfg.BlockStoreConfig.Directory)
			var bs blockstore.Blockstore
			bs = nil
			switch cfg.BlockStoreConfig.Type {
			case config.OptBadgerDS:
			case config.OptFlatFS:
				bs, err = ipfs.NewFlatfsBlockStore(bsDir)
				if err != nil {
					return err
				}
			}
			log.Info().Msg("Blockstore initialized")

			w, err := config.InitWallet(home)
			if err != nil {
				return err
			}
			log.Info().Str("provider_address", w.AccAddress()).Send()

			f, err := file_system.NewFileSystem(ctx, db, cfg.BlockStoreConfig.Key, ds, bs, cfg.APICfg.IPFSPort, cfg.APICfg.IPFSDomain)
			if err != nil {
				return err
			}

			unusedCids, err := f.ListUnusedCids(context.Background())
			if err != nil {
				return err
			}

			for _, cid := range unusedCids {
				fmt.Println(cid)
			}

			return nil
		},
	}
}
