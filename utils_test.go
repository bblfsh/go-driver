package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"github.com/src-d/babelfish-go-driver/msg"
)

type myTest struct {
	name string
	req  *msg.Request
	res  *msg.Response
}

// newMyTest creates a new test. It takes in the test name, the request for the input, the desired status response and the errors.
// If the desired status response is msg.Ok, the parameter errors must be nil.
func newMyTest(name string, req *msg.Request, status string, errors []string) *myTest {
	return &myTest{
		name: name,
		req:  req,
		res: &msg.Response{
			Status:          status,
			Errors:          errors,
			Driver:          driverVersion,
			Language:        lang,
			LanguageVersion: langVersion,
			AST:             getTree(req.Content),
		},
	}
}

// getTree get the ast from a source.
func getTree(source string) *ast.File {
	fset := token.NewFileSet()
	tree, _ := parser.ParseFile(fset, "source.go", source, parser.ParseComments)
	ast.Inspect(tree, setObjNil)

	return tree
}

// loadFile generates a msg.Request with the content from a file.
func loadFile(name string) *msg.Request {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	source, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return &msg.Request{
		Action:  msg.ParseAst,
		Content: string(source),
	}
}
