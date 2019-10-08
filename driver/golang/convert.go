package golang

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

var (
	ScopeType   = reflect.TypeOf((*ast.Scope)(nil))
	ObjectType  = reflect.TypeOf((*ast.Object)(nil))
	NodeType    = reflect.TypeOf((*ast.Node)(nil)).Elem()
	PosType     = reflect.TypeOf(token.Pos(0))
	UASTIntType = reflect.TypeOf(nodes.Int(0))
)

var typeNameToType = make(map[string]reflect.Type)

func init() {
	registerType("ArrayType", ast.ArrayType{})
	registerType("AssignStmt", ast.AssignStmt{})
	registerType("BadDecl", ast.BadDecl{})
	registerType("BadExpr", ast.BadExpr{})
	registerType("BadStmt", ast.BadStmt{})
	registerType("BasicLit", ast.BasicLit{})
	registerType("BinaryExpr", ast.BinaryExpr{})
	registerType("BlockStmt", ast.BlockStmt{})
	registerType("BranchStmt", ast.BranchStmt{})
	registerType("CallExpr", ast.CallExpr{})
	registerType("CaseClause", ast.CaseClause{})
	registerType("ChanType", ast.ChanType{})
	registerType("CommClause", ast.CommClause{})
	registerType("Comment", ast.Comment{})
	registerType("CommentGroup", ast.CommentGroup{})
	registerType("CompositeLit", ast.CompositeLit{})
	registerType("CommentMap", ast.CommentMap{})
	registerType("DeclStmt", ast.DeclStmt{})
	registerType("DeferStmt", ast.DeferStmt{})
	registerType("Ellipsis", ast.Ellipsis{})
	registerType("EmptyStmt", ast.EmptyStmt{})
	registerType("ExprStmt", ast.ExprStmt{})
	registerType("Field", ast.Field{})
	registerType("FieldList", ast.FieldList{})
	registerType("File", ast.File{})
	registerType("ForStmt", ast.ForStmt{})
	registerType("FuncDecl", ast.FuncDecl{})
	registerType("FuncLit", ast.FuncLit{})
	registerType("FuncType", ast.FuncType{})
	registerType("GenDecl", ast.GenDecl{})
	registerType("GoStmt", ast.GoStmt{})
	registerType("Ident", ast.Ident{})
	registerType("IfStmt", ast.IfStmt{})
	registerType("ImportSpec", ast.ImportSpec{})
	registerType("IncDecStmt", ast.IncDecStmt{})
	registerType("IndexExpr", ast.IndexExpr{})
	registerType("InterfaceType", ast.InterfaceType{})
	registerType("KeyValueExpr", ast.KeyValueExpr{})
	registerType("LabeledStmt", ast.LabeledStmt{})
	registerType("MapType", ast.MapType{})
	registerType("Object", ast.Object{})
	registerType("Package", ast.Package{})
	registerType("ParenExpr", ast.ParenExpr{})
	registerType("RangeStmt", ast.RangeStmt{})
	registerType("ReturnStmt", ast.ReturnStmt{})
	registerType("Scope", ast.Scope{})
	registerType("SelectorExpr", ast.SelectorExpr{})
	registerType("SelectStmt", ast.SelectStmt{})
	registerType("SendStmt", ast.SendStmt{})
	registerType("SliceExpr", ast.SliceExpr{})
	registerType("StarExpr", ast.StarExpr{})
	registerType("StructType", ast.StructType{})
	registerType("SwitchStmt", ast.SwitchStmt{})
	registerType("TypeAssertExpr", ast.TypeAssertExpr{})
	registerType("TypeSpec", ast.TypeSpec{})
	registerType("TypeSwitchStmt", ast.TypeSwitchStmt{})
	registerType("UnaryExpr", ast.UnaryExpr{})
	registerType("ValueSpec", ast.ValueSpec{})
}

// FuncVisitor is a simple implementation of AST visitor that is used to post-process AST tree after the conversion
type FuncVisitor func(node ast.Node)

func (f FuncVisitor) Visit(node ast.Node) (w ast.Visitor) {
	f(node)
	return f
}

func registerType(tp string, rt interface{}) {
	typeNameToType[tp] = reflect.TypeOf(rt)
}

// NodeToAST uast/nodes node object and converts it to ast.Node
func NodeToAST(n nodes.Node) ast.Node {
	// if we return nil pointer as interface it means that interface with nil pointer will be returned
	// Elem() returns interface from nil from nil pointer inside interface
	// then we cast interface to ast.Node
	res := nodeToAST(n, NodeType).Interface().(ast.Node)
	// after previous casts some AST nodes type is assigned to nil
	// thus we traverse over the AST node and change nil pointers to the pointers to empty objects
	ast.Walk(FuncVisitor(func(node ast.Node) {
		switch o := node.(type) {
		case *ast.FuncDecl:
			if o.Type == nil {
				o.Type = &ast.FuncType{
					Params: &ast.FieldList{},
				}
			}
		case *ast.FuncType:
			if o.Params == nil {
				o.Params = &ast.FieldList{}
			}
		case *ast.FuncLit:
			if o.Type == nil {
				o.Type = &ast.FuncType{
					Params: &ast.FieldList{},
				}
			}
		}

	}), res)
	return res
}

func nodeToAST(n nodes.Node, t reflect.Type) reflect.Value {
	// switch on node types(Obj, Arr etc)
	// Obj has @type that is used as a map key
	switch o := n.(type) {
	case nil:
		if t == reflect.TypeOf(&ast.FuncType{}) {
			return reflect.New(t.Elem())
		}
		return reflect.Zero(t)
	case nodes.Object:
		// get @type from Object and get typeNameToType value from this key
		tp, ok := typeNameToType[uast.TypeOf(o)]
		if !ok {
			panic("not supported: " + uast.TypeOf(o))
		}
		val := reflect.New(tp).Elem()

		// iterate over Object fields
		for k, v := range o {
			// skip system fields
			if strings.Contains(k, "@") {
				continue
			}

			// get structure type field descriptor
			field, ok := tp.FieldByName(k)
			if !ok {
				panic(k + " field not found")
			}
			// if field is anonymous then we have an embedded struct then len(field.Index) >= 1
			// as far as we explicitly know that AST does not have these structs, we can skip handling this case
			if field.Anonymous {
				panic(k + " is anonymous")
			}

			// recursively call nodeToAST until go type(goTypeVal) is obtained
			desiredType := field.Type

			var goTypeVal reflect.Value
			// if we deal with token then we get token type of the value
			if desiredType == reflect.TypeOf(token.Token(0)) {
				goTypeVal = reflect.ValueOf(tokens[string(v.(nodes.String))])
			} else {
				// we need to pass the desiredType here to have a type t to pass to case nodes.Array:
				goTypeVal = nodeToAST(v, desiredType)
			}

			// if desired type is pointer, set(returned) type should be the reference to goTypeVal
			if desiredType.Kind() == reflect.Ptr || desiredType.Kind() == reflect.Interface {
				isZero := goTypeVal.IsZero()
				if isZero {
					if v == nil {
						continue
					}
					goTypeVal = reflect.New(goTypeVal.Type())
				}
			}

			// this allows us to convert typed go types to go types
			// Examples:
			// 1) type Kind int -> int
			// 2) int -> type Kind int
			convertedVal := goTypeVal.Convert(desiredType)
			// set the resulting value field as convertedVal
			val.Field(field.Index[0]).Set(convertedVal)
		}
		return val.Addr()
	case nodes.Array:
		// note arrays are slices of interfaces []Node, thus we need to init val in a different way
		// t is the type of the field
		ln := len(o)
		val := reflect.MakeSlice(t, ln, ln)
		// type of slice element
		te := t.Elem()

		for i, n := range o {
			// slice element is passed alongside with type of slice element
			goTypeVal := nodeToAST(n, te)
			// if desired type is pointer, set(returned) type should be the reference to goTypeVal

			// in the case of slice of interface implementations, that contains non-pointer implementation that implements interface with pointer receiver
			// then we return pointer to that implementation
			if !goTypeVal.Type().ConvertibleTo(te) {
				goTypeVal = goTypeVal.Addr()
			}
			// set to slice
			val.Index(i).Set(goTypeVal)
		}
		return val
	default:
		// for
		// nodes.String
		// nodes.Int
		// nodes.Uint
		// nodes.Float
		// nodes.Bool
		// and others
		return reflect.ValueOf(o).Convert(t)
	}
}

// ValueToNode takes an AST node/value and converts it to a tree of uast types
// like Object and List. In this case we have a full control of json encoding
// and can annotate the tree with native AST type names.
func ValueToNode(v interface{}, fs *token.FileSet) (nodes.Node, error) {
	val, ok := v.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(v)
	}

	if !val.IsValid() {
		return nil, nil
	}
	t := val.Type()
	switch t {
	case PosType:
		p := convertPosition(val.Interface().(token.Pos), fs)
		return p.ToObject(), nil
	}
	switch t.Kind() {
	case reflect.Slice:
		if val.Len() == 0 {
			return nil, nil
		}
		arr := make(nodes.Array, 0, val.Len())
		for i := 0; i < val.Len(); i++ {
			el, err := ValueToNode(val.Index(i), fs)
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
			fv := val.Field(i)
			if f.Type == PosType {
				p := convertPosition(fv.Interface().(token.Pos), fs)
				pos[f.Name] = p
				continue
			} else if f.Type == ScopeType || f.Type == ObjectType {
				// do not follow scope and object pointers - need a graph structure for it
				continue
			}
			el, err := ValueToNode(fv, fs)
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
		if val.IsNil() {
			return nil, nil
		}
		o, err := ValueToNode(val.Elem(), fs)
		if err != nil {
			return nil, err
		}
		if m, ok := o.(nodes.Object); ok && val.Type().Implements(NodeType) {
			n := val.Interface().(ast.Node)
			pos := uast.PositionsOf(m)
			if pos == nil {
				pos = make(uast.Positions)
			}
			pos[uast.KeyStart] = convertPosition(n.Pos(), fs)
			pos[uast.KeyEnd] = convertPosition(n.End(), fs)

			m[uast.KeyPos] = pos.ToObject()
		}
		return o, nil
	}
	o := val.Interface()
	if s, ok := o.(interface {
		String() string
	}); ok {
		return nodes.String(s.String()), nil
	} else if t.ConvertibleTo(UASTIntType) {
		return val.Convert(UASTIntType).Interface().(nodes.Int), nil
	}
	return uast.ToNode(o)
}

func convertPosition(p token.Pos, fs *token.FileSet) uast.Position {
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
