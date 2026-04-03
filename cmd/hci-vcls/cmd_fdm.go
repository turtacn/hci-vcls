package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func newFdmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fdm",
		Short: "FDM commands",
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

	evalCmd := &cobra.Command{
		Use: "evaluate",
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterID, _ := cmd.Flags().GetString("cluster-id")
			if clusterID == "" {
				return fmt.Errorf("cluster-id is required")
			}

			resp, err := http.Get("http://localhost:8080/api/v1/degradation")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var data map[string]interface{}
			_ = json.Unmarshal(body, &data)

			fmt.Fprintf(cmd.OutOrStdout(), "Degradation Evaluation: %v\n", data["level"])
			return nil
		},
	}
	evalCmd.Flags().String("cluster-id", "", "Cluster ID")
	cmd.AddCommand(evalCmd)

	return cmd
}

// Personal.AI order the ending
