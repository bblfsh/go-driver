package main

import "go/ast"

// setObjNil looks for the elements that can't be serialized to JSON and set it to nil.
func setObjNil(node interface{}) {
	switch node.(type) {
	case *ast.BranchStmt:
		n := node.(*ast.BranchStmt)
		safeIdent(n.Label)
	case *ast.File:
		n := node.(*ast.File)
		safeImportSpecList(n.Imports)
		safeIdent(n.Name)
		safeScope(n.Scope)
		safeIdentList(n.Unresolved)
	case *ast.FuncDecl:
		n := node.(*ast.FuncDecl)
		safeIdent(n.Name)
		safeFieldList(n.Recv)
		safeFuncTye(n.Type)
	case *ast.FuncLit:
		n := node.(*ast.FuncLit)
		safeFuncTye(n.Type)
	case *ast.FuncType:
		n := node.(*ast.FuncType)
		safeFuncTye(n)
	case *ast.Ident:
		n := node.(*ast.Ident)
		safeIdent(n)
	case *ast.ImportSpec:
		n := node.(*ast.ImportSpec)
		safeImportSpec(n)
	case *ast.InterfaceType:
		n := node.(*ast.InterfaceType)
		safeFieldList(n.Methods)
	case *ast.LabeledStmt:
		n := node.(*ast.LabeledStmt)
		safeIdent(n.Label)
	case *ast.Package:
		n := node.(*ast.Package)
		n.Files = nil
		n.Imports = nil
		safeScope(n.Scope)
	case *ast.Scope:
		n := node.(*ast.Scope)
		safeScope(n)
	case *ast.SelectorExpr:
		n := node.(*ast.SelectorExpr)
		safeIdent(n.Sel)
	case *ast.StructType:
		n := node.(*ast.StructType)
		safeFieldList(n.Fields)
	case *ast.TypeSpec:
		n := node.(*ast.TypeSpec)
		safeIdent(n.Name)
	}
}

func safeIdent(node *ast.Ident) {
	if node == nil {
		return
	}
	node.Obj = nil
}

func safeIdentList(list []*ast.Ident) {
	for i := range list {
		safeIdent(list[i])
	}
}

func safeImportSpec(is *ast.ImportSpec) {
	safeIdent(is.Name)
}

func safeImportSpecList(list []*ast.ImportSpec) {
	for i := range list {
		safeImportSpec(list[i])
	}
}

func safeField(field *ast.Field) {
	safeIdentList(field.Names)
}

func safeFieldList(flist *ast.FieldList) {
	if flist == nil {
		return
	}

	for i := range flist.List {
		safeField(flist.List[i])
	}
}

func safeFuncTye(ftype *ast.FuncType) {
	safeFieldList(ftype.Params)
	safeFieldList(ftype.Results)
}

func safeScope(scope *ast.Scope) {
	scope.Objects = nil
	scope.Outer = nil
}
