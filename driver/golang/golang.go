package golang

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"

	"github.com/opentracing/opentracing-go"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
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
	return convValue(reflect.ValueOf(f), fs)
}

var (
	scopeType   = reflect.TypeOf((*ast.Scope)(nil))
	objectType  = reflect.TypeOf((*ast.Object)(nil))
	nodeType    = reflect.TypeOf((*ast.Node)(nil)).Elem()
	posType     = reflect.TypeOf(token.Pos(0))
	uastIntType = reflect.TypeOf(nodes.Int(0))
)

func convPos(p token.Pos, fs *token.FileSet) uast.Position {
	if !p.IsValid() {
		return uast.Position{}
	}
	pos := fs.Position(p)
	return uast.Position{
		Offset: uint32(pos.Offset),
		Line:   uint32(pos.Line),
		Col:    uint32(pos.Column),
	}
}

// convValue takes an AST node/value and converts it to a tree of uast types
// like Object and List. In this case we have a full control of json encoding
// and can annotate the tree with native AST type names.
func convValue(v reflect.Value, fs *token.FileSet) (nodes.Node, error) {
	if !v.IsValid() {
		return nil, nil
	}
	t := v.Type()
	switch t {
	case posType:
		p := convPos(v.Interface().(token.Pos), fs)
		return p.ToObject(), nil
	}
	switch t.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return nil, nil
		}
		arr := make(nodes.Array, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			el, err := convValue(v.Index(i), fs)
			if err != nil {
				return nil, err
			}
			arr = append(arr, el)
		}
		return arr, nil
	case reflect.Struct:
		m := make(nodes.Object, t.NumField())
		m[uast.KeyType] = nodes.String(t.Name()) // annotate nodes with type names
		pos := make(uast.Positions)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fv := v.Field(i)
			if f.Type == posType {
				p := convPos(fv.Interface().(token.Pos), fs)
				pos[f.Name] = p
				continue
			} else if f.Type == scopeType || f.Type == objectType {
				// do not follow scope and object pointers - need a graph structure for it
				continue
			}
			el, err := convValue(fv, fs)
			if err != nil {
				return nil, err
			}
			m[f.Name] = el
		}
		if len(pos) != 0 {
			m[uast.KeyPos] = pos.ToObject()
		}
		return m, nil
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}
		o, err := convValue(v.Elem(), fs)
		if err != nil {
			return nil, err
		}
		if m, ok := o.(nodes.Object); ok && v.Type().Implements(nodeType) {
			n := v.Interface().(ast.Node)
			pos := uast.PositionsOf(m)
			if pos == nil {
				pos = make(uast.Positions)
			}
			pos[uast.KeyStart] = convPos(n.Pos(), fs)
			pos[uast.KeyEnd] = convPos(n.End(), fs)

			m[uast.KeyPos] = pos.ToObject()
		}
		return o, nil
	}
	o := v.Interface()
	if s, ok := o.(interface {
		String() string
	}); ok {
		return nodes.String(s.String()), nil
	} else if t.ConvertibleTo(uastIntType) {
		return v.Convert(uastIntType).Interface().(nodes.Int), nil
	}
	return uast.ToNode(o)
}

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
