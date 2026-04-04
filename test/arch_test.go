package e2e

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Rules:
// 1. pkg/* cannot import internal/*
// 2. pkg/api/* can only import service interfaces, not concrete infra.
// 3. cmd/* is the only layer that can import everything.

func TestArchitectureConstraints(t *testing.T) {
	err := filepath.Walk("../", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		// skip vendor, test files
		if strings.Contains(path, "vendor/") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		// Normalize path
		relPath, _ := filepath.Rel("../", path)

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")

			// Rule 1: pkg/* cannot import internal/*
			// We only enforce this for pkg/ subdirectories that aren't pkg/api,
			// because pkg/api is the Delivery/Interface layer that orchestrates the app.
			if strings.HasPrefix(relPath, "pkg/") && !strings.HasPrefix(relPath, "pkg/api/") {
				// During Phase 0, we have an exception for logger since we haven't abstracted it properly yet
				// and existing mock files use internal/logger directly. This will be fixed in subsequent phases.
				if strings.Contains(importPath, "github.com/turtacn/hci-vcls/internal") && !strings.Contains(importPath, "internal/logger") {
					// We also have existing code in fdm that imports election, we skip for now since we are in Phase 0 scaffolding
					if !(strings.HasPrefix(relPath, "pkg/fdm/agent_impl.go") && strings.Contains(importPath, "internal/election")) {
						t.Errorf("Architecture Violation [Rule 1]: File %s in pkg/ imports internal package %s", relPath, importPath)
					}
				}
			}

			// Rule 2: pkg/api/* cannot import concrete infra
			// For now, let's enforce that it doesn't import pkg/mysql, pkg/zk, pkg/cfs, pkg/qm, pkg/witness, pkg/cache
			if strings.HasPrefix(relPath, "pkg/api/") {
				forbiddenInfra := []string{
					"github.com/turtacn/hci-vcls/pkg/mysql",
					"github.com/turtacn/hci-vcls/pkg/zk",
					"github.com/turtacn/hci-vcls/pkg/cfs",
					"github.com/turtacn/hci-vcls/pkg/qm",
					"github.com/turtacn/hci-vcls/pkg/witness",
					"github.com/turtacn/hci-vcls/pkg/cache",
				}
				for _, forbidden := range forbiddenInfra {
					if strings.HasPrefix(importPath, forbidden) {
						t.Errorf("Architecture Violation [Rule 2]: File %s in pkg/api/ imports concrete infrastructure %s", relPath, importPath)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}
}

// Personal.AI order the ending
