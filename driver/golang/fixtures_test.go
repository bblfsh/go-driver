package golang

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bblfsh/go-driver/driver/normalizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/sdk/driver"
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/yaml.v2"
)

const fixturesDir = "../../fixtures"

const writeYml = false

func readFixturesFile(t testing.TB, name string) string {
	data, err := ioutil.ReadFile(filepath.Join(fixturesDir, name))
	require.NoError(t, err)
	return string(data)
}

func writeFixturesFile(t testing.TB, name string, data string) {
	err := ioutil.WriteFile(filepath.Join(fixturesDir, name), []byte(data), 0644)
	require.NoError(t, err)
}

func deleteFixturesFile(name string) {
	os.Remove(filepath.Join(fixturesDir, name))
}

func TestFixturesNative(t *testing.T) {
	list, err := ioutil.ReadDir(fixturesDir)
	require.NoError(t, err)
	for _, ent := range list {
		if !strings.HasSuffix(ent.Name(), ".go") {
			continue
		}
		name := strings.TrimSuffix(ent.Name(), ".go")
		t.Run(name, func(t *testing.T) {
			code := readFixturesFile(t, name+".go")

			resp, err := NewDriver().Parse(&driver.InternalParseRequest{
				Content:  string(code),
				Encoding: driver.Encoding(protocol.UTF8),
			})
			require.NoError(t, err)

			if writeYml {
				ya, err := yaml.Marshal(resp.AST)
				require.NoError(t, err)
				writeFixturesFile(t, name+".go.native.yml", string(ya))
			}

			js, err := json.Marshal(resp.AST)
			require.NoError(t, err)

			exp := readFixturesFile(t, name+".go.native")
			got := (&protocol.NativeParseResponse{
				Response: protocol.Response{
					Status: protocol.Status(resp.Status),
					Errors: resp.Errors,
				},
				AST:      string(js),
				Language: "go",
			}).String()
			if !assert.ObjectsAreEqual(exp, got) {
				writeFixturesFile(t, name+".go.native2", got)
			} else {
				deleteFixturesFile(name + ".go.native2")
			}
			require.Equal(t, exp, got)
		})
	}
}

func TestFixturesUAST(t *testing.T) {
	list, err := ioutil.ReadDir(fixturesDir)
	require.NoError(t, err)
	for _, ent := range list {
		if !strings.HasSuffix(ent.Name(), ".go") {
			continue
		}
		name := strings.TrimSuffix(ent.Name(), ".go")
		t.Run(name, func(t *testing.T) {
			code := readFixturesFile(t, name+".go")

			req := &driver.InternalParseRequest{
				Content:  string(code),
				Encoding: driver.Encoding(protocol.UTF8),
			}

			resp, err := NewDriver().Parse(req)
			require.NoError(t, err)

			ast, err := uast.ToNode(resp.AST)
			require.NoError(t, err)

			tr := driver.Transforms{
				Native: normalizer.Native,
				Code:   normalizer.Code,
			}
			ua, err := tr.Do(driver.ModeAST, code, ast)
			require.NoError(t, err)

			if writeYml {
				ya, err := yaml.Marshal(ua)
				require.NoError(t, err)
				writeFixturesFile(t, name+".go.uast.yml", string(ya))
			}

			un, err := protocol.ToNode(ua)
			require.NoError(t, err)

			exp := readFixturesFile(t, name+".go.uast")
			got := (&protocol.ParseResponse{
				Response: protocol.Response{
					Status: protocol.Status(resp.Status),
					Errors: resp.Errors,
				},
				UAST:     un,
				Language: "go",
			}).String()
			if !assert.ObjectsAreEqual(exp, got) {
				writeFixturesFile(t, name+".go.uast2", got)
			} else {
				deleteFixturesFile(name + ".go.uast2")
			}
			require.Equal(t, exp, got)
		})
	}
}

func BenchmarkTransform(b *testing.B) {
	const name = "json"
	code := readFixturesFile(b, name+".go")
	req := &driver.InternalParseRequest{
		Content:  string(code),
		Encoding: driver.Encoding(protocol.UTF8),
	}
	srv := NewDriver()

	tr := driver.Transforms{
		Native: normalizer.Native,
		Code:   normalizer.Code,
	}
	resp, err := srv.Parse(req)
	if err != nil {
		b.Fatal(err)
	}
	rast, err := uast.ToNode(resp.AST)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ast := rast.Clone()

		ua, err := tr.Do(driver.ModeAST, code, ast)
		if err != nil {
			b.Fatal(err)
		}

		un, err := protocol.ToNode(ua)
		if err != nil {
			b.Fatal(err)
		}
		_ = un
	}
}
