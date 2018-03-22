package normalizer

import (
	"go/token"
	"strings"

	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
	. "gopkg.in/bblfsh/sdk.v1/uast/transformer"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer/positioner"
)

// Native is the of list `transformer.Transformer` to apply to a native AST.
// To learn more about the Transformers and the available ones take a look to:
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast/transformer
var Native = Transformers([][]Transformer{
	{
		// ResponseMetadata is a transform that trims response metadata from AST.
		//
		// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast#ResponseMetadata
		ResponseMetadata{
			TopLevelIsRootNode: false,
		},
	},
	// The main block of transformation rules.
	{Mappings(Annotations...)},
	{
		// RolesDedup is used to remove duplicate roles assigned by multiple
		// transformation rules.
		RolesDedup(),
	},
}...)

// Code is a special block of transformations that are applied at the end
// and can access original source code file. It can be used to improve or
// fix positional information.
//
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast/transformer/positioner
var Code = []CodeTransformer{
	positioner.NewFillLineColFromOffset(),
}

func astLeft(typ string, ast ObjectOp) ObjectOp {
	return ASTObjectLeft(typ, ast)
}

func astRight(typ string, norm ObjectOp, rop ArrayOp, roles ...role.Role) ObjectOp {
	if static := typeRoles[typ]; len(static) > 0 {
		roles = append([]role.Role{}, roles...)
		roles = append(roles, static...)
	}
	return ASTObjectRight(typ, norm, rop, roles...)
}

// mapAST is a helper for describing a single AST transformation for a given node type.
func mapAST(typ string, ast, norm ObjectOp, roles ...role.Role) Mapping {
	return mapASTCustom(typ, ast, norm, nil, roles...)
}

func mapASTCustom(typ string, ast, norm ObjectOp, rop ArrayOp, roles ...role.Role) Mapping {
	return ASTMap(typ,
		astLeft(typ, ast),
		astRight(typ, norm, rop, roles...),
	)
}

type objRoles map[string][]role.Role

func annotateType(typ string, fields objRoles, roles ...role.Role) Mapping {
	left := make(Obj, len(fields))
	right := make(Obj, len(fields))
	for name, roles := range fields {
		left[name] = ObjectRoles(name + "_var")
		right[name] = ObjectRoles(name+"_var", roles...)
	}
	return mapAST(typ, left, right, roles...)
}

func operator(field, vr string, lookup map[uast.Value]ArrayOp, roles ...role.Role) Field {
	roles = append([]role.Role{
		role.Expression, role.Operator,
	}, roles...)
	var opRoles Op = Roles(roles...)
	if lookup != nil {
		opRoles = AppendRoles(
			LookupArrOpVar(vr, lookup),
			roles...,
		)
	}
	return Field{Name: field, Op: Fields{
		{Name: uast.KeyType, Op: String(uast.TypeOperator)},
		{Name: uast.KeyToken, Op: Var(vr)},
		{Name: uast.KeyRoles, Op: opRoles},
	}}
}

func astFieldLeft() Op {
	return astLeft("Field", Obj{
		"Names": Each("field_name", ObjectRoles("field_names")),
		"Type":  ObjectRoles("field_type"),
	})
}

func astFieldRight(inherit bool, roles ...role.Role) Op {
	nameRoles := []role.Role{role.Name}
	typeRoles := []role.Role{role.Type}
	if inherit {
		n := len(roles)
		nameRoles = append(roles[:n:n], nameRoles...)
		typeRoles = append(roles[:n:n], typeRoles...)
	}
	return astRight("Field", Obj{
		"Names": Each("field_name", ObjectRoles("field_names", nameRoles...)),
		"Type":  ObjectRoles("field_type", typeRoles...),
	}, nil, roles...)
}

var (
	literalRoles  = make(map[uast.Value]ArrayOp)
	opRolesBinary = make(map[uast.Value]ArrayOp)
	opRolesUnary  = make(map[uast.Value]ArrayOp)
	opIncDec      = make(map[uast.Value]ArrayOp)
	opAssign      = make(map[uast.Value]ArrayOp)
	branchRoles   = make(map[uast.Value]ArrayOp)
)

func goTok(tok token.Token) uast.Value {
	return uast.String(tok.String())
}

func fillTokenToRolesMap(dst map[uast.Value]ArrayOp, src map[token.Token][]role.Role) {
	for tok, roles := range src {
		dst[goTok(tok)] = Roles(roles...)
	}
}

func init() {
	fillTokenToRolesMap(literalRoles, map[token.Token][]role.Role{
		token.STRING: {role.String},
		token.CHAR:   {role.Character},
		token.INT:    {role.Number},
		token.FLOAT:  {role.Number},
		token.IMAG:   {role.Incomplete}, // TODO: IMAG
	})
	fillTokenToRolesMap(opRolesBinary, map[token.Token][]role.Role{
		token.ADD: {role.Arithmetic, role.Add},
		token.SUB: {role.Arithmetic, role.Substract},
		token.MUL: {role.Arithmetic, role.Multiply},
		token.QUO: {role.Arithmetic, role.Divide},
		token.REM: {role.Arithmetic, role.Modulo},

		token.XOR:     {role.Bitwise, role.Xor},
		token.OR:      {role.Bitwise, role.Or},
		token.AND:     {role.Bitwise, role.And},
		token.SHL:     {role.Bitwise, role.LeftShift},
		token.SHR:     {role.Bitwise, role.RightShift},
		token.AND_NOT: {role.Bitwise, role.And, role.Negative},

		token.LOR:  {role.Boolean, role.Or},
		token.LAND: {role.Boolean, role.And},

		token.EQL: {role.Relational, role.Equal},
		token.NEQ: {role.Relational, role.Not, role.Equal},
		token.LSS: {role.Relational, role.LessThan},
		token.GTR: {role.Relational, role.GreaterThan},
		token.LEQ: {role.Relational, role.LessThanOrEqual},
		token.GEQ: {role.Relational, role.GreaterThanOrEqual},
	})
	fillTokenToRolesMap(opRolesUnary, map[token.Token][]role.Role{
		token.ADD: {role.Arithmetic, role.Positive},
		token.SUB: {role.Arithmetic, role.Negative},

		token.MUL: {role.Dereference},
		token.AND: {role.TakeAddress},

		token.XOR: {role.Bitwise, role.Negative},
		token.NOT: {role.Boolean, role.Negative},

		token.ARROW: {role.Incomplete},
	})
	fillTokenToRolesMap(opIncDec, map[token.Token][]role.Role{
		token.INC: {role.Increment},
		token.DEC: {role.Decrement},
	})
	fillTokenToRolesMap(opAssign, map[token.Token][]role.Role{
		token.ASSIGN: {},
		token.DEFINE: {role.Declaration},

		token.ADD_ASSIGN: {role.Operator, role.Arithmetic, role.Add},
		token.SUB_ASSIGN: {role.Operator, role.Arithmetic, role.Substract},
		token.MUL_ASSIGN: {role.Operator, role.Arithmetic, role.Multiply},
		token.QUO_ASSIGN: {role.Operator, role.Arithmetic, role.Divide},
		token.REM_ASSIGN: {role.Operator, role.Arithmetic, role.Modulo},

		token.OR_ASSIGN:      {role.Operator, role.Bitwise, role.Or},
		token.AND_ASSIGN:     {role.Operator, role.Bitwise, role.And},
		token.XOR_ASSIGN:     {role.Operator, role.Bitwise, role.Xor},
		token.SHL_ASSIGN:     {role.Operator, role.Bitwise, role.LeftShift},
		token.SHR_ASSIGN:     {role.Operator, role.Bitwise, role.RightShift},
		token.AND_NOT_ASSIGN: {role.Operator, role.Bitwise, role.And, role.Negative},
	})
	fillTokenToRolesMap(branchRoles, map[token.Token][]role.Role{
		token.CONTINUE:    {role.Continue},
		token.BREAK:       {role.Break},
		token.GOTO:        {role.Goto},
		token.FALLTHROUGH: {role.Incomplete}, // TODO: fallthrough
	})
}

func uncomment(s string) (string, error) {
	// Remove // and /*...*/ from comment nodes' tokens
	if strings.HasPrefix(s, "//") {
		s = s[2:]
	} else if strings.HasPrefix(s, "/*") {
		s = s[2 : len(s)-2]
	}
	return s, nil
}

func comment(s string) (string, error) {
	if strings.Contains(s, "\n") {
		return "/*" + s + "*/", nil
	}
	return "//" + s, nil
}

// Annotations is a list of individual transformations to annotate a native AST with roles.
var Annotations = []Mapping{
	// ObjectToNode defines how to normalize common fields of native AST
	// (like node type, token, positional information).
	//
	// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast#ObjectToNode
	ObjectToNode{
		InternalTypeKey: "type",
		OffsetKey:       "start",
		EndOffsetKey:    "end",
	}.Mapping(),

	annotateType("File", nil, role.File),

	annotateType("CommentGroup", nil, role.Comment, role.List),

	mapAST("Comment", Obj{
		"Text": StringConv(Var("text"), uncomment, comment),
	}, Obj{ // ->
		uast.KeyToken: Var("text"),
	}, role.Comment),

	annotateType("BadExpr", nil, role.Incomplete),

	mapAST("Ident", Obj{
		"Name": Var("name"),
	}, Obj{ // ->
		uast.KeyToken: Var("name"),
	}, role.Identifier),

	mapASTCustom("BasicLit", Obj{
		"Value": Var("val"),
		"Kind":  Var("kind"),
	}, Fields{ // ->
		{Name: "Kind", Op: Var("kind")},
		{Name: uast.KeyToken, Op: Var("val")},
	}, LookupArrOpVar("kind", literalRoles),
		role.Literal, role.Primitive),

	mapASTCustom("BinaryExpr", Obj{
		"X":  ObjectRoles("left"),
		"Y":  ObjectRoles("right"),
		"Op": Var("op"),
	}, Pre(Fields{ // ->
		operator("Op", "op", opRolesBinary, role.Binary),
	}, Obj{
		"X": ObjectRoles("left", role.Binary, role.Left),
		"Y": ObjectRoles("right", role.Binary, role.Right),
	}), LookupArrOpVar("op", opRolesBinary), role.Binary),

	mapASTCustom("UnaryExpr", Obj{
		"X":  Var("x"),
		"Op": Var("op"),
	}, Fields{ // ->
		operator("Op", "op", opRolesUnary, role.Unary),
		{Name: "X", Op: Var("x")},
	}, LookupArrOpVar("op", opRolesUnary), role.Unary),

	mapASTCustom("IncDecStmt", Obj{
		"X":   Var("x"),
		"Tok": Var("op"),
	}, Fields{ // ->
		operator("Op", "op", opIncDec, role.Unary),
		{Name: "X", Op: Var("x")},
	}, LookupArrOpVar("op", opIncDec), role.Unary),

	annotateType("BlockStmt", nil, role.Block, role.Scope),

	mapASTCustom("AssignStmt", Obj{
		"Lhs": Each("lhs", ObjectRoles("left")),
		"Rhs": Each("rhs", ObjectRoles("right")),
		"Tok": Var("op"),
	}, Fields{ // ->
		operator("Op", "op", opAssign, role.Operator, role.Assignment, role.Binary),
		{Name: "Lhs", Op: Each("lhs", ObjectRoles("left", role.Assignment, role.Binary, role.Left))},
		{Name: "Rhs", Op: Each("rhs", ObjectRoles("right", role.Assignment, role.Binary, role.Right))},
	}, LookupArrOpVar("op", opAssign),
		role.Assignment, role.Binary),

	mapAST("IfStmt", Obj{
		"Init": OptObjectRoles("init"),
		"Cond": ObjectRoles("cond"),
		"Body": ObjectRoles("body"),
		"Else": OptObjectRoles("else"),
	}, Obj{ // ->
		"Init": OptObjectRoles("init", role.If, role.Initialization),
		"Cond": ObjectRoles("cond", role.If, role.Condition),
		"Body": ObjectRoles("body", role.Then, role.Body),
		"Else": OptObjectRoles("else", role.Else),
	}, role.If),

	mapAST("SwitchStmt", Obj{
		"Init": OptObjectRoles("init"),
		"Body": ObjectRoles("body"),
	}, Obj{ // ->
		"Init": OptObjectRoles("init", role.Switch, role.Initialization),
		"Body": ObjectRoles("body", role.Switch, role.Body),
	}, role.Switch),

	mapAST("TypeSwitchStmt", Obj{
		"Init": OptObjectRoles("init"),
		"Body": ObjectRoles("body"),
	}, Obj{ // ->
		"Init": OptObjectRoles("init", role.Switch, role.Initialization),
		"Body": ObjectRoles("body", role.Switch, role.Body),
	}, role.Switch, role.Incomplete),

	annotateType("SelectStmt", objRoles{
		"Body": {role.Switch, role.Body},
	}, role.Switch, role.Incomplete),

	mapAST("ForStmt", Obj{
		"Init": OptObjectRoles("init"),
		"Cond": OptObjectRoles("cond"),
		"Post": OptObjectRoles("post"),
		"Body": ObjectRoles("body"),
	}, Obj{ // ->
		"Init": OptObjectRoles("init", role.For, role.Initialization),
		"Cond": OptObjectRoles("cond", role.For, role.Condition),
		"Body": ObjectRoles("body", role.For, role.Body),
		"Post": OptObjectRoles("post", role.For, role.Update),
	}, role.For),

	mapAST("RangeStmt", Obj{
		"Key":   OptObjectRoles("key"),
		"Value": OptObjectRoles("val"),
		"X":     Var("x"),
		"Body":  ObjectRoles("body"),
	}, Obj{ // ->
		"Key":   OptObjectRoles("key", role.For, role.Iterator, role.Key),
		"Value": OptObjectRoles("val", role.For, role.Iterator, role.Value),
		"X":     Var("x"),
		"Body":  ObjectRoles("body", role.For, role.Body),
	}, role.For, role.Iterator),

	mapASTCustom("BranchStmt", Obj{
		"Label": Var("label"),
		"Tok":   Var("tok"),
	}, Fields{ // ->
		{Name: "Tok", Op: Var("tok")},
		{Name: "Label", Op: Var("label")},
	}, LookupArrOpVar("tok", branchRoles)),

	mapAST("ImportSpec", Obj{
		"Name": OptObjectRoles("name"),
		"Path": ObjectRoles("path"),
	}, Obj{ // ->
		"Name": OptObjectRoles("name", role.Import, role.Alias),
		"Path": ObjectRoles("path", role.Import, role.Pathname),
	}, role.Import, role.Declaration),

	mapAST("ValueSpec", Obj{
		"Type": OptObjectRoles("type"),
	}, Obj{ // ->
		"Type": OptObjectRoles("type", role.Type),
	}, role.Declaration),

	annotateType("TypeSpec", objRoles{
		"Type": {role.Type},
	}, role.Declaration),

	annotateType("ArrayType", objRoles{
		"Elt": {role.Entry},
	}, role.Type, role.List),

	annotateType("MapType", objRoles{
		"Key":   {role.Key},
		"Value": {role.Entry},
	}, role.Type, role.Map),

	annotateType("FuncLit", objRoles{
		"Type": {role.Type},
		"Body": {role.Body},
	}),

	mapAST("StructType", Fields{
		{Name: "Fields", Op: Part("fields", Obj{
			"List": Each("field", astFieldLeft()),
		})},
	}, Fields{ // ->
		{Name: "Fields", Op: Part("fields", Obj{
			"List": Each("field", astFieldRight(false, role.Entry)),
		})},
	}, role.Type),

	mapAST("InterfaceType", Fields{
		{Name: "Methods", Op: Part("fields", Fields{
			RolesField("field-list"),
			{Name: "List", Op: Each("field", astFieldLeft())},
		})},
	}, Fields{ // ->
		{Name: "Methods", Op: Part("fields", Fields{
			RolesField("field-list", role.Function, role.List),
			{Name: "List", Op: Each("field",
				astFieldRight(false, role.Entry),
			)},
		})},
	}, role.Type, role.Incomplete),

	mapAST("FuncType", Fields{
		{Name: "Params", Op: Part("params", Fields{
			RolesField("params-list"),
			{Name: "List", Op: Each("param", astFieldLeft())},
		})},
		{Name: "Results", Op: Opt("has-res", Part("results", Fields{
			RolesField("results-list"),
			{Name: "List", Op: Each("result", astFieldLeft())},
		}))},
	}, Fields{ // ->
		{Name: "Params", Op: Part("params", Fields{
			RolesField("params-list", role.ArgsList),
			{Name: "List", Op: Each("param",
				astFieldRight(false, role.Argument),
			)},
		})},
		{Name: "Results", Op: Opt("has-res", Part("results", Fields{
			RolesField("results-list", role.Return, role.ArgsList),
			{Name: "List", Op: Each("result",
				astFieldRight(false, role.Return, role.Argument),
			)},
		}))},
	}, role.Function, role.Type),

	mapAST("FuncDecl", Fields{
		{Name: "Recv", Op: Opt("recv_set", Part("recv", Fields{
			RolesField("field-list"),
			{Name: "List", Op: Each("field", astFieldLeft())},
		}))},
		{Name: "Name", Op: ObjectRoles("name")},
		{Name: "Type", Op: ObjectRoles("type")},
		{Name: "Body", Op: ObjectRoles("body")},
	}, Fields{ // ->
		{Name: "Recv", Op: Opt("recv_set", Part("recv", Fields{
			RolesField("field-list", role.Function, role.Receiver, role.List),
			{Name: "List", Op: Each("field",
				astFieldRight(true, role.Function, role.Receiver),
			)},
		}))},
		{Name: "Name", Op: ObjectRoles("name", role.Function, role.Name)},
		{Name: "Type", Op: ObjectRoles("type", role.Function, role.Type)},
		{Name: "Body", Op: ObjectRoles("body", role.Function, role.Body)},
	}, role.Function, role.Declaration),

	mapAST("GenDecl", Fields{
		{Name: "Tok", Op: Is(goTok(token.VAR))},
		{Name: "Specs", Op: Each("specs", Part("spec", Obj{
			"Names": Each("names", ObjectRoles("var_names")),
		}))},
	}, Fields{ // ->
		{Name: "Tok", Op: Is(goTok(token.VAR))},
		{Name: "Specs", Op: Each("specs", Part("spec", Obj{
			"Names": Each("names", ObjectRoles("var_names", role.Variable, role.Name)),
		}))},
	}, role.Variable, role.Declaration),

	mapAST("GenDecl", Fields{
		{Name: "Tok", Op: Is(goTok(token.CONST))},
		{Name: "Specs", Op: Each("specs", Part("spec", Obj{
			"Names": Each("names", ObjectRoles("const_names")),
		}))},
	}, Fields{ // ->
		{Name: "Tok", Op: Is(goTok(token.CONST))},
		{Name: "Specs", Op: Each("specs", Part("spec", Obj{
			"Names": Each("names", ObjectRoles("const_names", role.Name)),
		}))},
	}, role.Incomplete, role.Declaration),

	mapAST("GenDecl", Fields{
		{Name: "Tok", Op: Is(goTok(token.TYPE))},
		{Name: "Specs", Op: Each("specs", Part("spec", Obj{
			"Name": ObjectRoles("type_name"),
		}))},
	}, Fields{ // ->
		{Name: "Tok", Op: Is(goTok(token.TYPE))},
		{Name: "Specs", Op: Each("specs", Part("spec", Obj{
			"Name": ObjectRoles("type_name", role.Type, role.Name),
		}))},
	}, role.Type, role.Declaration),

	mapAST("GenDecl", Fields{
		{Name: "Tok", Op: Is(goTok(token.IMPORT))},
	}, Fields{ // ->
		{Name: "Tok", Op: Is(goTok(token.IMPORT))},
	}, role.Declaration),

	mapAST("CallExpr", Obj{
		"Fun":  ObjectRoles("func"),
		"Args": Each("args", ObjectRoles("arg")),
	}, Obj{ // ->
		"Fun":  ObjectRoles("func", role.Callee),
		"Args": Each("args", ObjectRoles("arg", role.Argument, role.Positional)),
	}, role.Call),

	annotateType("KeyValueExpr", objRoles{
		"Key":   {role.Key},
		"Value": {role.Value},
	}),

	annotateType("CaseClause", nil, role.Case),
	annotateType("CommClause", nil, role.Case, role.Incomplete),
	// TODO: default (length of List is zero)

	annotateType("ReturnStmt", nil, role.Return),
	annotateType("GoStmt", nil, role.Incomplete),       // TODO: Async role
	annotateType("SelectorExpr", nil, role.Incomplete), // TODO: new role

	annotateType("CompositeLit", nil, role.Literal),
	annotateType("ChanType", nil, role.Type, role.Incomplete),

	annotateType("ExprStmt", nil),
	annotateType("DeclStmt", nil),
	annotateType("DeferStmt", nil, role.Incomplete),
	annotateType("SendStmt", nil, role.Incomplete),
	annotateType("LabeledStmt", nil),
	annotateType("Ellipsis", nil),
	annotateType("SliceExpr", nil),
	annotateType("IndexExpr", nil),
	annotateType("StarExpr", nil),
	annotateType("TypeAssertExpr", nil, role.Incomplete),
}

/*

	// Declarations
	On(goast.GenDecl).Roles(role.Declaration).Self(
		On(HasProperty("Tok", "var")).Roles(role.Variable).Children(
			OnIntRole("Specs").Children(
				OnIntRole("Names", role.Variable, role.Name),
			),
		),
		// TODO: Constant role
		On(HasProperty("Tok", "const")).Roles(role.Incomplete).Children(
			OnIntRole("Specs").Children(
				OnIntRole("Names", role.Name),
			),
		),
		On(HasProperty("Tok", "type")).Roles(role.Type).Children(
			OnIntRole("Specs").Children(
				OnIntRole("Names", role.Type, role.Name),
			),
		),
	),

*/
