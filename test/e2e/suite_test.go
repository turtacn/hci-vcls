package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/turtacn/hci-vcls/test/e2e/helpers"
)

var (
	testCluster *helpers.TestCluster
)

func TestMain(m *testing.M) {
	fmt.Println("Setting up E2E test suite...")
	testCluster = helpers.NewTestCluster()

	err := testCluster.Start()
	if err != nil {
		fmt.Printf("Failed to start test cluster: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	fmt.Println("Tearing down E2E test suite...")
	err = testCluster.Stop()
	if err != nil {
		fmt.Printf("Failed to stop test cluster: %v\n", err)
	}

	os.Exit(code)
}

//Personal.AI order the ending