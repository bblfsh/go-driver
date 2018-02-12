package normalizer

import (
	"gopkg.in/bblfsh/sdk.v1/uast"
)

// ToNode is an instance of `uast.ObjectToNode`, defining how to transform an
// into a UAST (`uast.Node`).
//
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast#ObjectToNode
var ToNode = &uast.ObjectToNode{
	InternalTypeKey: "type",
	TokenKeys: map[string]bool{
		"Value": true,
		"Name":  true,
		"Text":  true,
	},
	PromotedPropertyStrings: map[string]map[string]bool{
		"BinaryExpr": {"Op": true},
	},
}
