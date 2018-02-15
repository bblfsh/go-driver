package normalizer

import (
	"errors"

	"github.com/bblfsh/go-driver/driver/goast"
	"gopkg.in/bblfsh/sdk.v1/uast"
	. "gopkg.in/bblfsh/sdk.v1/uast/ann"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer/annotatter"
)

// Transformers is the of list `transformer.Transfomer` to apply to a UAST, to
// learn more about the Transformers and the available ones take a look to:
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast/transformer
var Transformers = []transformer.Tranformer{
	annotatter.NewAnnotatter(AnnotationRulesExpr),
	annotatter.NewAnnotatter(AnnotationRulesStmt),
	annotatter.NewAnnotatter(AnnotationRules),
}

func OnIntRole(r string, roles ...uast.Role) *Rule {
	return On(HasInternalRole(r)).Roles(roles...)
}

func OnIntType(it string, roles ...uast.Role) *Rule {
	return On(HasInternalType(it)).Roles(roles...)
}

// AnnotationRules describes how a UAST should be annotated with `uast.Role`.
//
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast/ann
var AnnotationRules = On(Any).Self(
	On(Not(goast.File)).Error(errors.New("root must be uast.File")),
	On(goast.File).Roles(uast.File).Descendants(
		// Identifiers
		On(goast.Ident).Roles(uast.Identifier),

		// Literals
		On(goast.BasicLit).Roles(uast.Literal, uast.Primitive).Self(
			On(Or(
				HasProperty("Kind", "INT"),
				HasProperty("Kind", "FLOAT"),
			)).Roles(uast.Number),
			On(HasProperty("Kind", "STRING")).Roles(uast.String),
			On(HasProperty("Kind", "CHAR")).Roles(uast.Character),
			// TODO: IMAG
		),

		// Binary Expression
		On(goast.BinaryExpr).Roles(uast.Binary).Children(
			OnIntType("BinaryExpr.Op", uast.Expression, uast.Binary, uast.Operator).Self(
				On(HasToken("+")).Roles(uast.Arithmetic, uast.Add),
				On(HasToken("-")).Roles(uast.Arithmetic, uast.Substract),
				On(HasToken("*")).Roles(uast.Arithmetic, uast.Multiply),
				On(HasToken("/")).Roles(uast.Arithmetic, uast.Divide),
				On(HasToken("%")).Roles(uast.Arithmetic, uast.Modulo),

				On(HasToken("^")).Roles(uast.Bitwise, uast.Xor),
				On(HasToken("|")).Roles(uast.Bitwise, uast.Or),
				On(HasToken("&")).Roles(uast.Bitwise, uast.And),
				On(HasToken("<<")).Roles(uast.Bitwise, uast.LeftShift),
				On(HasToken(">>")).Roles(uast.Bitwise, uast.RightShift),

				On(HasToken("||")).Roles(uast.Boolean, uast.Or),
				On(HasToken("&&")).Roles(uast.Boolean, uast.And),

				On(HasToken("==")).Roles(uast.Relational, uast.Equal),
				On(HasToken("!=")).Roles(uast.Relational, uast.Not, uast.Equal),
				On(HasToken("<")).Roles(uast.Relational, uast.LessThan),
				On(HasToken(">")).Roles(uast.Relational, uast.GreaterThan),
				On(HasToken("<=")).Roles(uast.Relational, uast.LessThanOrEqual),
				On(HasToken(">=")).Roles(uast.Relational, uast.GreaterThanOrEqual),
			),
			OnIntRole("X", uast.Binary, uast.Left),
			OnIntRole("Y", uast.Binary, uast.Right),
		),

		// Unary Expression
		On(goast.UnaryExpr).Roles(uast.Unary).Children(
			OnIntType("UnaryExpr.Op", uast.Expression, uast.Unary, uast.Operator).Self(
				On(HasToken("+")).Roles(uast.Arithmetic, uast.Positive),
				On(HasToken("-")).Roles(uast.Arithmetic, uast.Negative),
				On(HasToken("*")).Roles(uast.Dereference),
				On(HasToken("&")).Roles(uast.TakeAddress),
				On(HasToken("^")).Roles(uast.Bitwise, uast.Negative),
				On(HasToken("!")).Roles(uast.Boolean, uast.Negative),
			),
		),

		// Inc/Dec statement
		On(goast.IncDecStmt).Self(
			On(HasToken("--")).Roles(uast.Decrement),
			On(HasToken("++")).Roles(uast.Increment),
		),

		// Assignment
		On(goast.AssignStmt).Roles(uast.Assignment).Self(
			On(HasToken("+=")).Roles(uast.Operator, uast.Binary, uast.Arithmetic, uast.Add),
			On(HasToken("-=")).Roles(uast.Operator, uast.Binary, uast.Arithmetic, uast.Substract),
			On(HasToken("*=")).Roles(uast.Operator, uast.Binary, uast.Arithmetic, uast.Multiply),
			On(HasToken("/=")).Roles(uast.Operator, uast.Binary, uast.Arithmetic, uast.Divide),
			On(HasToken("%=")).Roles(uast.Operator, uast.Binary, uast.Arithmetic, uast.Modulo),

			On(HasToken("|=")).Roles(uast.Operator, uast.Binary, uast.Bitwise, uast.Or),
			On(HasToken("&=")).Roles(uast.Operator, uast.Binary, uast.Bitwise, uast.And),
			On(HasToken("^=")).Roles(uast.Operator, uast.Binary, uast.Bitwise, uast.Xor),
			On(HasToken("<<=")).Roles(uast.Operator, uast.Binary, uast.Bitwise, uast.LeftShift),
			On(HasToken(">>=")).Roles(uast.Operator, uast.Binary, uast.Bitwise, uast.RightShift),
		).Children(
			OnIntRole("Lhs", uast.Assignment, uast.Binary, uast.Left),
			OnIntRole("Rhs", uast.Assignment, uast.Binary, uast.Right),
		),

		// Comments
		On(goast.CommentGroup).Roles(uast.Comment, uast.List),
		On(goast.Comment).Roles(uast.Comment),

		// Blocks
		On(goast.BlockStmt).Roles(uast.Block, uast.Scope).Self(
			OnIntRole("Body", uast.Body),
		),

		// If statement
		On(goast.IfStmt).Roles(uast.If).Children(
			OnIntRole("Cond", uast.If, uast.Condition),
			OnIntRole("Body", uast.Then),
			OnIntRole("Else", uast.Else),
		),

		// For loop
		On(goast.ForStmt).Roles(uast.For).Children(
			OnIntRole("Cond", uast.For, uast.Condition),
			OnIntRole("Init", uast.For, uast.Initialization),
			OnIntRole("Post", uast.For, uast.Update),
			OnIntRole("Body", uast.For, uast.Body),
		),

		// For range
		On(goast.RangeStmt).Roles(uast.For, uast.Iterator).Children(
			OnIntRole("Key", uast.For, uast.Iterator, uast.Key),
			OnIntRole("Value", uast.For, uast.Iterator, uast.Value),
			OnIntRole("Body", uast.For, uast.Body),
		),

		// Branch statements
		On(goast.BranchStmt).Self(
			On(HasProperty("Tok", "continue")).Roles(uast.Continue),
			On(HasProperty("Tok", "break")).Roles(uast.Break),
			On(HasProperty("Tok", "goto")).Roles(uast.Goto),
			// TODO: fallthrough
		),

		// Switch statement
		On(goast.SwitchStmt).Roles(uast.Switch),
		On(goast.CaseClause).Roles(uast.Case),
		// TODO: default (length of List is zero)

		// Declarations
		On(goast.GenDecl).Roles(uast.Declaration).Self(
			On(HasProperty("Tok", "var")).Roles(uast.Variable).Children(
				OnIntRole("Specs").Children(
					OnIntRole("Names", uast.Variable, uast.Name),
				),
			),
			// TODO: Constant role
			On(HasProperty("Tok", "const")).Roles(uast.Incomplete).Children(
				OnIntRole("Specs").Children(
					OnIntRole("Names", uast.Name),
				),
			),
			On(HasProperty("Tok", "type")).Roles(uast.Type).Children(
				OnIntRole("Specs").Children(
					OnIntRole("Names", uast.Type, uast.Name),
				),
			),
		),

		// Imports
		On(goast.ImportSpec).Roles(uast.Import, uast.Declaration).Children(
			OnIntRole("Name", uast.Import, uast.Alias),
			OnIntRole("Path", uast.Import, uast.Pathname),
		),

		// Var/Const declarations (see GenDecl)
		On(goast.ValueSpec).Roles(uast.Declaration).Children(
			OnIntRole("Type", uast.Type),
		),

		// Type declarations
		On(goast.TypeSpec).Roles(uast.Declaration).Children(
			OnIntRole("Type", uast.Type),
		),

		// Arrays and slices
		On(goast.ArrayType).Roles(uast.Type, uast.List).Children(
			OnIntRole("Elt", uast.Entry),
		),

		// Maps
		On(goast.MapType).Roles(uast.Type, uast.Map).Children(
			OnIntRole("Key", uast.Key),
			OnIntRole("Value", uast.Entry),
		),

		// Channels
		On(goast.ChanType).Roles(uast.Type, uast.Incomplete), // TODO: channels

		// Structs
		On(goast.StructType).Roles(uast.Type).Children(
			OnIntRole("Fields").Children(
				On(goast.Field).Roles(uast.Entry).Children(
					OnIntRole("Names").Roles(uast.Name),
					OnIntRole("Type").Roles(uast.Type),
				),
			),
		),

		// Interfaces
		On(goast.InterfaceType).Roles(uast.Type, uast.Incomplete).Children(
			OnIntRole("Methods", uast.Function, uast.List).Children(
				On(goast.Field).Roles(uast.Entry).Children(
					OnIntRole("Names").Roles(uast.Name),
					OnIntRole("Type").Roles(uast.Type),
				),
			),
		),

		// Function type
		On(goast.FuncType).Roles(uast.Function, uast.Type).Children(
			On(goast.FieldList).Roles(uast.ArgsList).Children(
				On(goast.Field).Roles(uast.Argument).Children(
					OnIntRole("Names").Roles(uast.Name),
					OnIntRole("Type").Roles(uast.Type),
				),
			),
		),

		// Function declaration
		On(goast.FuncDecl).Roles(uast.Declaration, uast.Function).Children(
			OnIntRole("Recv", uast.Function, uast.Receiver, uast.List).Children(
				On(goast.Field).Roles(uast.Function, uast.Receiver).Children(
					OnIntRole("Names", uast.Function, uast.Receiver, uast.Name),
					OnIntRole("Type", uast.Function, uast.Receiver, uast.Type),
				),
			),
			OnIntRole("Name").Roles(uast.Function, uast.Name),
			OnIntRole("Type").Roles(uast.Function, uast.Type),
			OnIntRole("Body").Roles(uast.Function),
		),

		// Function calls
		On(goast.CallExpr).Roles(uast.Call).Children(
			OnIntRole("X", uast.Receiver),
			OnIntRole("Fun", uast.Callee),
			OnIntRole("Args", uast.Argument, uast.Positional),
		),

		// Return
		On(goast.ReturnStmt).Roles(uast.Return),

		// Goroutine
		On(goast.GoStmt).Roles(uast.Incomplete), // TODO: Async role

		// Field access
		On(goast.SelectorExpr).Roles(uast.Incomplete), // TODO: new role

		// Composite literals
		On(goast.CompositeLit).Roles(uast.Literal),
		On(goast.KeyValueExpr).Children(
			OnIntRole("Key", uast.Key),
			OnIntRole("Value", uast.Value),
		),
	),
)
