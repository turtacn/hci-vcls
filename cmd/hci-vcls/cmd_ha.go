package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newHACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ha",
		Short: "HA commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use: "tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "HA Tasks")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use: "evaluate",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "HA Evaluate")
			return nil
		},
	})

	return cmd
}

// Personal.AI order the ending