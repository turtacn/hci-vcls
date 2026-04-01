package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFdmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fdm",
		Short: "FDM commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "FDM Status")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use: "degradation",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "FDM Degradation")
			return nil
		},
	})

	return cmd
}

// Personal.AI order the ending