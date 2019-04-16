package golang

import (
	"context"
	"testing"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/stretchr/testify/require"
)

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
