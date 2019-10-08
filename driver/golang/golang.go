package golang

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"

	"github.com/bblfsh/go-driver/driver/golang/convert"

	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/opentracing/opentracing-go"
)

func ParseString(code string) (*ast.File, *token.FileSet, error) {
	fs := token.NewFileSet()
	tree, err := parser.ParseFile(fs, "input.go", code, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return tree, fs, nil
}

func Parse(code string) (nodes.Node, error) {
	f, fs, err := ParseString(code)
	if err != nil {
		return nil, err
	}
	return convert.ValueToNode(reflect.ValueOf(f), fs)
}

var (
	scopeType   = reflect.TypeOf((*ast.Scope)(nil))
	objectType  = reflect.TypeOf((*ast.Object)(nil))
	nodeType    = reflect.TypeOf((*ast.Node)(nil)).Elem()
	posType     = reflect.TypeOf(token.Pos(0))
	uastIntType = reflect.TypeOf(nodes.Int(0))
)

func NewDriver() *Driver {
	return &Driver{}
}

type Driver struct{}

func (Driver) Start() error {
	return nil
}
func (Driver) Close() error {
	return nil
}
func (Driver) Parse(ctx context.Context, code string) (nodes.Node, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "go.Parse")
	defer sp.Finish()

	return Parse(code)
}
