package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version command",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get("http://localhost:8080/api/v1/version")
		if err != nil {
			fmt.Printf("Version: %s, Commit: %s, Date: %s\n", Version, Commit, Date)
			return nil
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

