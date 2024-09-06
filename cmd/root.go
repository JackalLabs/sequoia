package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/JackalLabs/jackal-provider/jprov/crypto"
	"github.com/JackalLabs/jackal-provider/jprov/utils"
	"github.com/cosmos/cosmos-sdk/client"
	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/cmd/wallet"
	"github.com/JackalLabs/sequoia/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Level(zerolog.InfoLevel)
}

func InitCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "init",
		Short: "initializes sequoias config folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			_, err = config.Init(home)
			if err != nil {
				return err
			}
			_, err = config.InitWallet(home)
			if err != nil {
				return err
			}

			fmt.Println("done!")

			return nil
		},
	}

	return r
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("failed to read string")
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func ShutdownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Permanently remove provider from network and get deposit back",
		Long:  "Permanently remove provider from network and get deposit back.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if !askForConfirmation("Terminate Provider Permanently?") {
				return nil
			}

			if !askForConfirmation("You're absolutely sure?") {
				return nil
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				fmt.Println(err)
				return err
			}

			address, err := crypto.GetAddress(clientCtx)
			if err != nil {
				fmt.Println(err)
				return err
			}

			fmt.Printf("Terminating provider: %s\n", address)
			msg := storageTypes.NewMsgShutdownProvider(
				address,
			)
			if err := msg.ValidateBasic(); err != nil {
				fmt.Println(err)
				return err
			}
			res, err := utils.SendTx(clientCtx, cmd.Flags(), "", msg)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println(res.RawLog)
			return err
		},
	}

	return cmd
}

func VersionCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "version",
		Short: "checks the version of sequoia",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\nCommit: %s\n", config.Version(), config.Commit())

			return nil
		},
	}

	return r
}

func RootCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "sequoia",
		Short: "Sequoia is a fast and light-weight Jackal Storage Provider.",
	}

	r.PersistentFlags().String(types.FlagHome, types.DefaultHome, "sets the home directory for sequoia")
	r.PersistentFlags().String(types.FlagLogLevel, types.DefaultLogLevel, "log level. info|error|debug")

	r.AddCommand(StartCmd(), wallet.WalletCmd(), InitCmd(), VersionCmd(), IPFSCmd(), SalvageCmd(), ShutdownCmd())

	return r
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {

		log.Error().Err(err)
		os.Exit(1)
	}
}
