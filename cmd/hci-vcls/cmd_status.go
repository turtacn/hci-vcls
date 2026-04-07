package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Status command",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get("http://localhost:8080/api/v1/status")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintln(cmd.OutOrStdout(), string(body))
			return nil
		},
	}
}

