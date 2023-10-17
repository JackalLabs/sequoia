package cmd

import (
	"fmt"
	"github.com/JackalLabs/sequoia/cmd/wallet"
	"github.com/spf13/cobra"
	"os"
)

func RootCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "sequoia",
		Short: "Sequoia is a fast and light-weight Jackal Storage Provider.",
	}

	r.AddCommand(StartCmd(), InitCmd(), wallet.WalletCmd())

	return r
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
