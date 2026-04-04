package main

import "os"

// Version information set by ldflags
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}

// Personal.AI order the ending
