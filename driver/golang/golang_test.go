package golang

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/protocol"
	"gopkg.in/bblfsh/sdk.v2/sdk/driver"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func TestNative(t *testing.T) {
	const code = `package main`

	resp, err := NewDriver().Parse(&driver.InternalParseRequest{
		Encoding: driver.Encoding(protocol.UTF8),
		Content:  string(code),
	})
	require.NoError(t, err)
	require.Equal(t, driver.Status(protocol.Ok), resp.Status)
	require.Empty(t, resp.Errors)
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
	require.Equal(t, exp, resp.AST)
}
