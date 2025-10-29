package database

import (
	"fmt"
	"os"
	"runtime"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/config"
	"github.com/JackalLabs/sequoia/utils"
	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/cobra"
)

func DataCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "data",
		Short: "Data subcommands",
	}

	c.AddCommand(keysCmd(), getObjectCmd(), garbageCmd())

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
