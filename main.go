package main

import (
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/src-d/lang-parsers/go/go-driver/msg"

	"github.com/ugorji/go/codec"
)

const (
	lang = "Go"
)

var (
	langVersion   = runtime.Version()
	driverVersion string
)

func main() {
	in := os.Stdin
	out := os.Stdout

	if err := start(in, out); err != nil {
		log.Fatal(err)
	}
}

// start launchs a loop to read requests and write responses.
func start(in io.Reader, out io.Writer) error {
	var mpHandle codec.MsgpackHandle
	mpDec := codec.NewDecoder(in, &mpHandle)
	mpEnc := codec.NewEncoder(out, &mpHandle)
	req := &msg.Request{}
	var res *msg.Response

	for {
		if err := mpDec.Decode(req); err != nil {
			if err == io.EOF {
				break
			}

			res = &msg.Response{
				Status:          msg.Fatal,
				Errors:          []string{err.Error()},
				Language:        lang,
				LanguageVersion: langVersion,
				Driver:          driverVersion,
			}
			mpEnc.MustEncode(res)
			return err
		}

		res = getResponse(req)
		mpEnc.MustEncode(res)
	}

	return nil
}

// getResponse always generates a msg.Response. The response will have the properly status (Ok, Error, Fatal).
func getResponse(m *msg.Request) *msg.Response {
	res := &msg.Response{
		Language:        lang,
		LanguageVersion: langVersion,
		Driver:          driverVersion,
	}

	fset := token.NewFileSet()
	tree, err := parser.ParseFile(fset, "source.go", m.Content, parser.ParseComments|parser.AllErrors)
	if err != nil {
		if tree == nil {
			res.Status = msg.Fatal
			res.Errors = []string{err.Error()}
			return res
		}

		res.Status = msg.Error
		errList := err.(scanner.ErrorList)
		res.Errors = getErrors(errList)
	} else {
		res.Status = msg.Ok
	}

	ast.Inspect(tree, setObjNil)
	res.AST = tree

	return res
}

// getErrors build a []string with the err.Error() from a scanner.ErrorList.
func getErrors(errList scanner.ErrorList) []string {
	list := make([]string, 0, len(errList))
	for _, err := range errList {
		list = append(list, err.Error())
	}

	return list
}
