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
	OffsetKey:    "offset",
	EndOffsetKey: "end",

	Modifier: func(n map[string]interface{}) error {
		rename := func(from, to string) {
			if v, ok := n[from]; ok {
				delete(n, from)
				n[to] = v
			}
		}
		switch t, _ := n["type"].(string); t {
		case "Ident":
			rename("NamePos", "offset")
		case "BasicLit":
			rename("ValuePos", "offset")
		case "GenDecl":
			rename("TokPos", "offset")
			rename("Rparen", "end")
		case "BlockStmt":
			rename("Lbrace", "offset")
			rename("Rbrace", "end")
		case "FuncType":
			rename("Func", "offset")
		case "DeferStmt":
			rename("Defer", "offset")
		case "Comment":
			if text, ok := n["Text"].(string); ok {
				// Remove // and /*...*/ from comment nodes' tokens
				if strings.HasPrefix(text, "//") {
					n["Text"] = text[2:]
				} else if strings.HasPrefix(text, "/*") {
					n["Text"] = text[2 : len(text)-2]
				}
			}
			rename("Slash", "offset")
		}
		return nil
	},
}
