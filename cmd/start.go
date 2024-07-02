package cmd

import (
	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func StartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts the provider",
		Run: func(cmd *cobra.Command, args []string) {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				panic(err)
			}

			logLevel, err := cmd.Flags().GetString(types.FlagLogLevel)
			if err != nil {
				panic(err)
			}

			if logLevel == "info" {
				log.Logger = log.Level(zerolog.InfoLevel)
			} else if logLevel == "debug" {
				log.Logger = log.Level(zerolog.DebugLevel)
			} else if logLevel == "error" {
				log.Logger = log.Level(zerolog.ErrorLevel)
			}

			app := core.NewApp(home)

			app.Start()
		},
	}
}
