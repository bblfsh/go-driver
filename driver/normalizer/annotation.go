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

		// Comments
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
					OnIntRole("Names", uast.Variable),
				),
			),
			// TODO: constants
			//On(HasProperty("Tok","const")).Roles(uast.Constant).Children(
			//	OnIntRole("Specs", uast.Constant).Children(
			//		OnIntRole("Names", uast.Constant),
			//	),
			//),
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

		// Function calls
		On(goast.CallExpr).Roles(uast.Call).Children(
			OnIntRole("X", uast.Receiver),
			OnIntRole("Fun", uast.Callee),
			OnIntRole("Args", uast.Argument, uast.Positional),
		),
	),
)
