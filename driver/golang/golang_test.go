package golang

import (
	"bytes"
	"context"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"

	"github.com/stretchr/testify/require"
)

const (
	fixtures               = "../../fixtures"
	exclusionFileSubstring = "error"
)

var whiteList = []string{
	"augmented_assign.go",
	"bench_binary_search.go",
	"bench_fibonacci.go",
	"bench_fizzbuzz.go",
	"bench_palindrome.go",
	"comparison.go",
	"increment.go",
	"pointers.go",
	"primitives.go",
	"recover.go",
	"u2_class_field.go",
	"u2_class_field_annotation.go",
	"u2_class_field_qualifiers.go",
	"u2_func_return_multiple.go",
	"u2_func_simple.go",
	"u2_import_module_alias.go",
	"u2_import_path.go",
	"u2_import_specific_init.go",
	"u2_import_subsymbols_namespaced.go",
	"u2_type_interface.go",
	"u2_type_qualifiers.go",
	"unicode.go",
}

func TestNative(t *testing.T) {
	const code = `package main`

	ast, err := NewDriver().Parse(context.Background(), code)
	require.NoError(t, err)
	pos := func(p, l, c int) nodes.Node {
		return uast.Position{
			Offset: uint32(p),
			Line:   uint32(l),
			Col:    uint32(c),
		}.ToObject()
	}
	type str = nodes.String

	exp := nodes.Object{
		uast.KeyType: str("File"),
		uast.KeyPos: nodes.Object{
			uast.KeyType:  str(uast.TypePositions),
			uast.KeyStart: pos(0, 1, 1),
			uast.KeyEnd:   pos(12, 1, 13),
			"Package":     pos(0, 1, 1),
		},
		"Name": nodes.Object{
			uast.KeyType: str("Ident"),
			uast.KeyPos: nodes.Object{
				uast.KeyType:  str(uast.TypePositions),
				uast.KeyStart: pos(8, 1, 9),
				uast.KeyEnd:   pos(12, 1, 13),
				"NamePos":     pos(8, 1, 9),
			},
			"Name": str("main"),
		},
		"Imports":    nil,
		"Comments":   nil,
		"Doc":        nil,
		"Decls":      nil,
		"Unresolved": nil,
	}
	require.Equal(t, exp, ast)
}

// TestStringReplace tests simple replace string in code transformation
// NOTE! due to positions issue previous and final strings have the same length
// NOTE! tested code is sensible to blocks' empty lines, because we still do not preserve the positions
func TestStringReplaceTransform(t *testing.T) {
	in, err := Parse(getCode("foo"))
	require.NoError(t, err)

	expCode := getCode("bar")
	exp, err := Parse(expCode)
	require.NoError(t, err)

	m := transformer.Mappings(
		transformer.Map(
			transformer.String("\"foo\""),
			transformer.String("\"bar\""),
		),
	)

	act, err := m.Do(in)
	require.NoError(t, err)
	require.Equal(t, exp, act)

	buf := &bytes.Buffer{}
	require.NoError(t, printer.Fprint(buf, token.NewFileSet(), NodeToAST(act)))
	require.Equal(t, expCode, buf.String(), buf.String())
}

// TestUASTNodeToCode
// 1) parse code ${exp_code} to UAST node
// 2) convert UAST node to AST node
// 3) convert AST node to ${act_code}
// <expected> formatted(${exp_code}) eq formatted(${act_code})
// TODO(lwsanty): research for approaches to preserve positions, see https://github.com/dave/dst
// NOTE! currently positions are not preserved during the conversions, that leads to the cases when we cannot preserve
// the code-style(not fmt-related issue!) that leads to some fixtures failure.
// Ones that pass are whitelisted in testdata/whitelist.yml
func TestUASTNodeToCode(t *testing.T) {
	files, err := selectFiles()
	require.NoError(t, err)
	t.Logf("matches: %v", files)

	wl, err := getWhiteListMap()
	require.NoError(t, err)

	for _, f := range files {
		fBase := filepath.Base(f)
		t.Run(fBase, func(t *testing.T) {
			if _, ok := wl[fBase]; !ok {
				t.Skipf("skipping test for %s: not supported yet", fBase)
			}

			data, err := ioutil.ReadFile(f)
			require.NoError(t, err)

			expCode, err := formatCode(string(data))
			require.NoError(t, err)
			node, err := Parse(expCode)
			require.NoError(t, err)

			actCode, err := nodeToCode(node)
			require.NoError(t, err)
			require.Equal(t, expCode, actCode)
		})
	}
}

func nodeToCode(n nodes.Node) (string, error) {
	astNode := NodeToAST(n)

	buf := &bytes.Buffer{}
	if err := format.Node(buf, token.NewFileSet(), astNode); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func formatCode(code string) (string, error) {
	fSet := token.NewFileSet()
	node, err := parser.ParseFile(fSet, "test.go", code, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = format.Node(&buf, fSet, node)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func selectFiles() ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(fixtures, "*.go"))
	if err != nil {
		return nil, err
	}

	for i, m := range matches {
		if !strings.Contains(m, exclusionFileSubstring) {
			continue
		}
		matches = append(matches[:i], matches[i+1:]...)
	}
	return matches, nil
}

// aaaa
func getCode(name string) string {
	return `package main

import "fmt"

func main() {
	var a string = "` + name + `"
	yo := func() {
		fmt.Println("zdarov")
	}
	yo()
	fmt.Println(a)
}
`
}

func getWhiteListMap() (map[string]struct{}, error) {
	res := make(map[string]struct{})
	for _, f := range whiteList {
		res[f] = struct{}{}
	}
	return res, nil
}
