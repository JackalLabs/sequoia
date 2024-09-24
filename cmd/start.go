package cmd

import (
	"fmt"
	"time"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultMaxRestartAttempt = 60

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			logLevel, err := cmd.Flags().GetString(types.FlagLogLevel)
			if err != nil {
				return err
			}

			maxRestartAttempt, err := cmd.Flags().GetInt("restart-attempt")
			if err != nil {
				return err
			}
			if maxRestartAttempt < 0 {
				maxRestartAttempt = 0
			}

			if logLevel == "info" {
				log.Logger = log.Level(zerolog.InfoLevel)
			} else if logLevel == "debug" {
				log.Logger = log.Level(zerolog.DebugLevel)
			} else if logLevel == "error" {
				log.Logger = log.Level(zerolog.ErrorLevel)
			}

			app, err := core.NewApp(home)
			if err != nil {
				return err
			}

			err = app.Start()

			for restartAttempt := 0; restartAttempt < maxRestartAttempt && err != nil; restartAttempt++ {
				fmt.Println(err)
				fmt.Printf("Attempting restart again in a minute (attempt %d of %d)...\n", restartAttempt+1, maxRestartAttempt)
				time.Sleep(time.Minute)
				err = app.Start()
			}
			return nil
		},
	}

	cmd.Flags().Int("restart-attempt", defaultMaxRestartAttempt, "attempt to restart <restart-attempt> times when the provider fails to start")
	cmd.Flags().String("ip", "http://example.com", "provider comain")
	cmd.Flags().Int64("apicfg.port", 3333, "port to serve api requests")
	cmd.Flags().Int("apicfg.ipfsport", 4005, "port for IPFS")
	cmd.Flags().String("apicfg.ipfsdomain", "dns4/ipfs.example.com/tcp/4001", "IPFS domain")
	cmd.Flags().Int64("proofthreads", 1000, "maximum threads for proofs")
	cmd.Flags().String("datadirectory", "$HOME/.sequoia/data", "directory to store database files")
	cmd.Flags().Int64("queueinterval", 10, "seconds to wait until next cycle to flush the transaction queue")
	cmd.Flags().Int64("proofinterval", 120, "seconds to wait until next cycle to post proofs")
	cmd.Flags().Int64("totalspace", 1092616192, "maximum storage space to provide in bytes")

	viper.BindPFlags(cmd.Flags())

	return cmd
}
