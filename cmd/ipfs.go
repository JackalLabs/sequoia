package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func IPFSCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "ipfs",
		Short: "Details about the IPFS network",
	}
	c.AddCommand(PeersCmd())
	return c
}

func PeersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peers",
		Short: "list peers on ipfs",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get("http://localhost:3333/ipfs/peers")
			if err != nil {
				return err
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			fmt.Println(string(body))
			return nil
		},
	}
}
