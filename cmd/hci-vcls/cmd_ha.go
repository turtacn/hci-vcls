package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func newHACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ha",
		Short: "HA commands",
	}

	cmd.PersistentFlags().String("cluster-id", "", "Cluster ID")

	cmd.AddCommand(&cobra.Command{
		Use: "tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get("http://localhost:8080/api/v1/ha/tasks")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintln(cmd.OutOrStdout(), string(body))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use: "evaluate",
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterID, _ := cmd.Flags().GetString("cluster-id")
			if clusterID == "" {
				return fmt.Errorf("cluster-id is required")
			}

			reqBody, _ := json.Marshal(map[string]string{"cluster_id": clusterID})

			resp, err := http.Post("http://localhost:8080/api/v1/ha/evaluate", "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintln(cmd.OutOrStdout(), string(body))
			return nil
		},
	})

	return cmd
}

