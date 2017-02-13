package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	"github.com/src-d/lang-parsers/go/go-driver/msg"
	"github.com/ugorji/go/codec"
)

var (
	mpHandle codec.MsgpackHandle
	mpDec    *codec.Decoder
	mpEnc    *codec.Encoder

	in  = os.Stdin
	out = os.Stdout

	req = &msg.Request{}
	res *msg.Response
)

func init() {
	mpDec = codec.NewDecoder(in, &mpHandle)
	mpEnc = codec.NewEncoder(out, &mpHandle)
}

func main() {
	mpDec.MustDecode(req)
	res, err := getResponse(req)
	if err != nil {
		log.Fatal(err)
	}
	mpEnc.MustEncode(res)
}

func getResponse(m *msg.Request) (*msg.Response, error) {
	fset := token.NewFileSet()
	tree, err := parser.ParseFile(fset, "source.go", m.Content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Inspect(tree, func(n ast.Node) bool {
		setObjNil(n)
		return true
	})

	res := &msg.Response{
		Status: msg.Ok,
		AST:    tree,
	}

	return res, nil
}
