package fixtures

import (
	"path/filepath"
	"testing"

	"github.com/bblfsh/go-driver/driver/golang"
	"github.com/bblfsh/go-driver/driver/normalizer"
	"gopkg.in/bblfsh/sdk.v2/sdk/driver"
	"gopkg.in/bblfsh/sdk.v2/sdk/driver/fixtures"
)

const projectRoot = "../../"

var Suite = &fixtures.Suite{
	Lang: "go",
	Ext:  ".go",
	Path: filepath.Join(projectRoot, fixtures.Dir),
	NewDriver: func() driver.BaseDriver {
		return golang.NewDriver()
	},
	Transforms: normalizer.Transforms,
	BenchName:  "json",
	Semantic: fixtures.SemanticConfig{
		BlacklistTypes: []string{
			"Ident",
			"Comment",
			"BlockStmt",
			"ImportSpec",
			"FuncDecl",
			"FuncType",
		},
	},
	Docker: fixtures.DockerConfig{
		Image: "golang:1.10",
	},
}

func TestGoDriver(t *testing.T) {
	Suite.RunTests(t)
}

func BenchmarkGoDriver(b *testing.B) {
	Suite.RunBenchmarks(b)
}
