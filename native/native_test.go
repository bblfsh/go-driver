package main

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/sdk/driver"
)

func TestNative(t *testing.T) {
	const code = `package main`

	resp := NewServer().handle(driver.InternalParseRequest{
		Encoding: driver.Encoding(protocol.UTF8),
		Content:  string(code),
	})
	require.Equal(t, driver.Status(protocol.Ok), resp.Status)
	require.Empty(t, resp.Errors)
	exp := astRoot{
		AST: map[string]interface{}{
			"Package": token.Pos(1), "type": "File",
			"Name": map[string]interface{}{
				"type": "Ident", "NamePos": token.Pos(9), "Name": "main",
			},
		},
	}
	require.Equal(t, exp, resp.AST)
}
