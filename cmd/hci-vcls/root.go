package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/turtacn/hci-vcls/internal/config"
)

var (
	cfgFile   string
	appConfig *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "hci-vcls",
	Short: "HCI vCLS Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("HCI vCLS Service")
		fmt.Printf("Version: %s, Commit: %s, Date: %s\n", BuildVersion, BuildCommit, BuildDate)
		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hci-vcls.yaml)")

	rootCmd.AddCommand(newFdmCmd())
	rootCmd.AddCommand(newHACmd())
	rootCmd.AddCommand(newStatusCmd())
}

func initConfig() {
	var err error
	appConfig, err = config.Load(cfgFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// Personal.AI order the ending