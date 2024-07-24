package cmd

import (
	"fmt"
	"time"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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
	return cmd
}
