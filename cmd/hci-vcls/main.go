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

//Personal.AI order the ending