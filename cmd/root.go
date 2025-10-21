package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"

	storageTypes "github.com/jackalLabs/canine-chain/v5/x/storage/types"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/cmd/wallet"
	"github.com/JackalLabs/sequoia/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		switch response {
		case "y", "yes":
			return true
		case "n", "no":
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
			if !askForConfirmation("Terminate Provider Permanently?") || !askForConfirmation("You're absolutely sure?") {
				fmt.Println("Provider termination cancelled")
				return nil
			}

			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			_, err = config.Init(home)
			if err != nil {
				return err
			}

			wallet, err := config.InitWallet(home)
			if err != nil {
				return err
			}

			fmt.Printf("Terminating provider: %s\n", wallet.AccAddress())
			msg := storageTypes.NewMsgShutdownProvider(
				wallet.AccAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				fmt.Println(err)
				return err
			}

			data := walletTypes.NewTransactionData(
				msg,
			).WithGasAuto().WithFeeAuto()

			res, err := wallet.BroadcastTxCommit(data)
			if err != nil {
				return err
			}

			if res.Code == 0 {
				fmt.Println("Shutdown successful!")
			} else {
				fmt.Println("Something went wrong, please try again.")
			}
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

// RootCmd creates and returns the root Cobra command for the Sequoia CLI, configuring global flags and adding all subcommands.
func RootCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "sequoia",
		Short: "Sequoia is a fast and light-weight Jackal Storage Provider.",
	}

	r.PersistentFlags().String(types.FlagHome, types.DefaultHome, "sets the home directory for sequoia")
	r.PersistentFlags().String(types.FlagLogLevel, types.DefaultLogLevel, "log level. info|error|debug")
	r.PersistentFlags().Int("restart-attempt", defaultMaxRestartAttempt, "attempt to restart <restart-attempt> times when the provider fails to start")
	r.PersistentFlags().String("domain", "http://example.com", "provider domain")
	r.PersistentFlags().Int64("api_config.port", 3333, "port to serve api requests")
	r.PersistentFlags().Int("api_config.ipfs_port", 4005, "port for IPFS")
	r.PersistentFlags().String("api_config.ipfs_domain", "dns4/ipfs.example.com/tcp/4001", "IPFS domain")
	r.PersistentFlags().Bool("api_config.ipfs_search", true, "Search for IPFS connections on Jackal on startup")
	r.PersistentFlags().Bool("api_config.open_gateway", true, "Open gateway for file retrieval even if file is on a different provider")
	r.PersistentFlags().Int64("proof_threads", 1000, "maximum threads for proofs")
	r.PersistentFlags().String("data_directory", "$HOME/.sequoia/data", "directory to store database files")
	r.PersistentFlags().Int64("queue_interval", 10, "seconds to wait until next cycle to flush the transaction queue")
	r.PersistentFlags().Int64("proof_interval", 120, "seconds to wait until next cycle to post proofs")
	r.PersistentFlags().Int64("total_bytes_offered", 1092616192, "maximum storage space to provide in bytes")

	err := viper.BindPFlags(r.PersistentFlags())
	if err != nil {
		panic(err)
	}

	r.AddCommand(StartCmd(), wallet.WalletCmd(), InitCmd(), VersionCmd(), IPFSCmd(), ShutdownCmd())

	return r
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {

		log.Error().Err(err)
		os.Exit(1)
	}
}
