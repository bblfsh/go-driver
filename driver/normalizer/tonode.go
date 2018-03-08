package normalizer

import (
	"strings"

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

	Modifier: func(n map[string]interface{}) error {
		// Remove // and /*...*/ from comment nodes' tokens
		if t, ok := n["type"].(string); ok && t == "Comment" {
			if text, ok := n["Text"].(string); ok {
				if strings.HasPrefix(text, "//") {
					n["Text"] = text[2:];
				} else if strings.HasPrefix(text, "/*") {
					n["Text"] = text[2:len(text)-2]
				}
			}
		}
		return nil
	},
}
