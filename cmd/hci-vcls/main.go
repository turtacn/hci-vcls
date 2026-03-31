package main

import (
	"fmt"
	"os"
)

var (
	BuildVersion string
	BuildCommit  string
	BuildDate    string
)

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Execute is a placeholder for the root command execution.
// It will be replaced by the cobra root command in cmd/hci-vcls/root.go
func Execute() error {
	fmt.Println("HCI vCLS Service")
	fmt.Printf("Version: %s, Commit: %s, Date: %s\n", BuildVersion, BuildCommit, BuildDate)
	return nil
}

//Personal.AI order the ending