package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	"go/ast"
	"go/parser"
	"go/token"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/sdk/driver"
	"gopkg.in/bblfsh/sdk.v1/sdk/jsonlines"
)

const TypeField = "type"

func NewServer() *Server {
	return &Server{}
}

type Server struct{}

func ParseString(code string) (*ast.File, error) {
	fs := token.NewFileSet()
	tree, err := parser.ParseFile(fs, "input.go", code, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func errResp(err error) driver.InternalParseResponse {
	return driver.InternalParseResponse{
		Status: driver.Status(protocol.Fatal),
		Errors: []string{err.Error()},
	}
}

func errRespf(format string, args ...interface{}) driver.InternalParseResponse {
	return errResp(fmt.Errorf(format, args...))
}

type astRoot struct {
	AST interface{} `json:"GOAST"`
}

func (s *Server) handle(req driver.InternalParseRequest) driver.InternalParseResponse {
	code, err := req.Encoding.Decode(req.Content)
	if err != nil {
		return errRespf("failed to decode contents: %v", err)
	}
	f, err := ParseString(code)
	if err != nil {
		return errResp(err)
	}
	return driver.InternalParseResponse{
		Status: driver.Status(protocol.Ok),
		AST: astRoot{
			AST: convValue(reflect.ValueOf(f)),
		},
	}
}

var (
	scopeType  = reflect.TypeOf((*ast.Scope)(nil))
	objectType = reflect.TypeOf((*ast.Object)(nil))
	tokenType  = reflect.TypeOf(token.Token(0))
	nodeType   = reflect.TypeOf((*ast.Node)(nil)).Elem()
)

// convValue takes an AST node/value and converts it to a tree of generic types
// like map[string]interface{} and []interface{}. In this case we have a full control
// of json encoding and can annotate the tree with native AST type names.
func convValue(v reflect.Value) interface{} {
	if !v.IsValid() {
		return nil
	}
	t := v.Type()
	if t == tokenType {
		return v.Interface().(token.Token).String()
	}
	switch t.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		var arr []interface{}
		for i := 0; i < v.Len(); i++ {
			arr = append(arr, convValue(v.Index(i)))
		}
		return arr
	case reflect.Struct:
		m := make(map[string]interface{}, t.NumField())
		m[TypeField] = t.Name() // annotate nodes with type names
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Type == scopeType || f.Type == objectType {
				// do not follow scope and object pointers - need a graph structure for it
				continue
			}
			m[f.Name] = convValue(v.Field(i))
		}
		return m
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil
		}
		o := convValue(v.Elem())
		if m, ok := o.(map[string]interface{}); ok && v.Type().Implements(nodeType) {
			n := v.Interface().(ast.Node)
			m["start"] = n.Pos()
			m["end"] = n.End()
		}
		return o
	}
	return v.Interface()
}

func (s *Server) Serve(c io.ReadWriter) error {
	enc := jsonlines.NewEncoder(c)
	dec := jsonlines.NewDecoder(c)
	for {
		var req driver.InternalParseRequest
		err := dec.Decode(&req)
		if err == io.EOF {
			return nil
		} else if err != nil {
			err = enc.Encode(errRespf("failed to decode request: %v", err))
			if err != nil {
				return err
			}
			continue
		}
		resp := s.handle(req)
		if err = enc.Encode(resp); err != nil {
			return err
		}
	}
}

func main() {
	srv := NewServer()
	c := struct {
		io.Reader
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}
	if err := srv.Serve(c); err != nil {
		log.Fatal(err)
	}
}
