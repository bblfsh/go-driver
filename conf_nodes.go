package main

import "go/ast"

// setObjNil looks for the elements that can't be serialized and set it to nil.
// It has the properly signature to be a parameter of ast.Inspect function.
func setObjNil(node ast.Node) bool {
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

	return true
}

// safeIdent set to nil conflictives fields of a ast.Ident.
func safeIdent(node *ast.Ident) {
	if node == nil {
		return
	}

	node.Obj = nil
}

// safeIdentList iterates over a slice of ast.Ident and calls safeIdent.
func safeIdentList(list []*ast.Ident) {
	for i := range list {
		safeIdent(list[i])
	}
}

// safeImportSpec set to nil conflictives fields of a ast.ImportSpect.
func safeImportSpec(is *ast.ImportSpec) {
	safeIdent(is.Name)
}

// safeImportSpecList iterates over a slice of ast.ImportSpec and calls safeImportSpec.
func safeImportSpecList(list []*ast.ImportSpec) {
	for i := range list {
		safeImportSpec(list[i])
	}
}

// safeField set to nil conflictives fields of a ast.Field.
func safeField(field *ast.Field) {
	safeIdentList(field.Names)
}

// safeFieldList iterates over a slice of ast.Field and calls safeField.
func safeFieldList(flist *ast.FieldList) {
	if flist == nil {
		return
	}

	for i := range flist.List {
		safeField(flist.List[i])
	}
}

// safeFuncTye set to nil conflictives fields of a ast.FuncType.
func safeFuncTye(ftype *ast.FuncType) {
	safeFieldList(ftype.Params)
	safeFieldList(ftype.Results)
}

// safeScope set to nil conflictives fields of a ast.Scope.
func safeScope(scope *ast.Scope) {
	scope.Objects = nil
	scope.Outer = nil
}
