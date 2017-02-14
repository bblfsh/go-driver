package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/src-d/lang-parsers/go/go-driver/msg"
	"github.com/ugorji/go/codec"
)

var (
	source = `package main
import "fmt"

func main() {
	fmt.Println("Hello World!")
}
`
	req1 = &msg.Request{
		Action:  msg.ParseAst,
		Content: source,
	}
	res1 = &msg.Response{
		Status:          msg.Ok,
		Driver:          driverVersion,
		Language:        language,
		LanguageVersion: languageVersion,
		AST:             getTree(req1.Content),
	}

	req2 = &msg.Request{
		Action: msg.ParseAst,
	}
	res2 = &msg.Response{
		Status:          msg.Error,
		Errors:          []string{"source.go:1:1: expected 'package', found 'EOF'"},
		Driver:          driverVersion,
		Language:        language,
		LanguageVersion: languageVersion,
	}

	req3 = loadFile("testfiles/test2.go")
	res3 = &msg.Response{
		Status:          msg.Ok,
		Driver:          driverVersion,
		Language:        language,
		LanguageVersion: languageVersion,
		AST:             getTree(req3.Content),
	}

	req4 = loadFile("testfiles/test3.go")
	res4 = &msg.Response{
		Status:          msg.Ok,
		Driver:          driverVersion,
		Language:        language,
		LanguageVersion: languageVersion,
		AST:             getTree(req4.Content),
	}

	req5 = loadFile("testfiles/test4.go")
	res5 = &msg.Response{
		Status:          msg.Ok,
		Driver:          driverVersion,
		Language:        language,
		LanguageVersion: languageVersion,
		AST:             getTree(req5.Content),
	}
)

func Test_getResponse(t *testing.T) {
	type args struct {
		m *msg.Request
	}
	tests := []struct {
		name string
		args args
		want *msg.Response
	}{
		{
			name: "statusOK",
			args: args{m: req1},
			want: res1,
		},
		{
			name: "statusFatal",
			args: args{m: req2},
			want: res2,
		},
		{
			name: "test2.go",
			args: args{m: req3},
			want: res3,
		},
		{
			name: "test3.go",
			args: args{m: req4},
			want: res4,
		},
		{
			name: "test4.go",
			args: args{m: req5},
			want: res5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getResponse(tt.args.m)
			if !resEquals(got, tt.want) {
				t.Errorf("getResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

// loadFile generates a msg.Request with content from a file.
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

// getTree get the ast from a source.
func getTree(source string) *ast.File {
	fset := token.NewFileSet()
	tree, _ := parser.ParseFile(fset, "source.go", source, parser.ParseComments)
	return tree
}

// resEquals compare two Responses.
func resEquals(got, want *msg.Response) bool {
	if got.Status != want.Status {
		return false
	}

	if len(got.Errors) != len(want.Errors) {
		return false
	}

	for i, v := range got.Errors {
		if v != want.Errors[i] {
			return false
		}
	}

	if got.Driver != want.Driver {
		return false
	}

	if got.Language != want.Language {
		return false
	}

	if got.LanguageVersion != want.LanguageVersion {
		return false
	}

	if got.LanguageVersion != want.LanguageVersion {
		return false
	}

	if got.Status == msg.Ok && !equalsMsgpack(got.AST, want.AST) {
		return false
	}

	return true
}

// equalsMsgpack serializes two AST and compare them.
func equalsMsgpack(gotTree, wantTree *ast.File) bool {
	// prepare to serialize
	ast.Inspect(gotTree, setObjNil)
	ast.Inspect(wantTree, setObjNil)

	// get io.Writers to serialize
	gotBuf := &bytes.Buffer{}
	wantBuf := &bytes.Buffer{}

	// get a Messagepack handle
	var mpHandler codec.MsgpackHandle
	mpHandler.Canonical = true

	// get encoders
	gotEnc := codec.NewEncoder(gotBuf, &mpHandler)
	wantEnc := codec.NewEncoder(wantBuf, &mpHandler)

	// encode trees
	gotEnc.MustEncode(gotTree)
	wantEnc.MustEncode(wantTree)

	// compare both serializations
	return bytes.Equal(gotBuf.Bytes(), wantBuf.Bytes())
}
