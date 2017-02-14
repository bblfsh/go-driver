package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/src-d/lang-parsers/go/go-driver/msg"

	"github.com/ugorji/go/codec"
)

const (
	language = "Go"
)

var (
	languageVersion = runtime.Version()
	driverVersion   string
)

func main() {
	in := os.Stdin
	out := os.Stdout

	if err := start(in, out); err != nil {
		log.Fatal(err)
	}
}

func start(in io.Reader, out io.Writer) error {
	var mpHandle codec.MsgpackHandle
	mpDec := codec.NewDecoder(in, &mpHandle)
	mpEnc := codec.NewEncoder(out, &mpHandle)
	req := &msg.Request{}

	for {
		if err := mpDec.Decode(req); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		res := getResponse(req)
		mpEnc.MustEncode(res)
	}
	return nil
}

// getResponse always generates a msg.Response. The response will have the properly status (Ok, Error, Fatal).
func getResponse(m *msg.Request) *msg.Response {
	res := &msg.Response{
		Language:        language,
		LanguageVersion: languageVersion,
		Driver:          driverVersion,
	}

	fset := token.NewFileSet()
	tree, err := parser.ParseFile(fset, "source.go", m.Content, parser.ParseComments)
	if err != nil {
		res.Errors = []string{err.Error()}
		if tree == nil {
			res.Status = msg.Fatal
			return res
		}

		res.Status = msg.Error
	} else {
		res.Status = msg.Ok
	}

	ast.Inspect(tree, setObjNil)

	res.AST = tree
	return res
}
