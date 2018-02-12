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
			OnIntType("BinaryExpr.Op", uast.Expression, uast.Binary, uast.Operator),
			OnIntRole("X", uast.Binary, uast.Left),
			OnIntRole("Y", uast.Binary, uast.Right),
		),

		// Unary Expression
		On(goast.UnaryExpr).Roles(uast.Unary).Children(
			OnIntType("UnaryExpr.Op", uast.Expression, uast.Unary, uast.Operator),
		),

		// Comments
		On(goast.Comment).Roles(uast.Comment),

		// Blocks
		On(goast.BlockStmt).Roles(uast.Block).Self(
			On(HasInternalRole("Body")).Roles(uast.Body),
		),

		// If statement
		On(goast.IfStmt).Roles(uast.If).Children(
			On(HasInternalRole("Cond")).Roles(uast.If, uast.Condition),
			On(HasInternalRole("Body")).Roles(uast.Then),
			On(HasInternalRole("Else")).Roles(uast.Else),
		),

		// For loop
		On(goast.ForStmt).Roles(uast.For).Children(
			On(HasInternalRole("Cond")).Roles(uast.For, uast.Condition),
			On(HasInternalRole("Body")).Roles(uast.For, uast.Body),
		),

		// Branch statements
		On(goast.BranchStmt).Self(
			On(HasProperty("Tok", "continue")).Roles(uast.Continue),
			On(HasProperty("Tok", "break")).Roles(uast.Break),
			On(HasProperty("Tok", "goto")).Roles(uast.Goto),
			// TODO: fallthrough
		),
	),
)
