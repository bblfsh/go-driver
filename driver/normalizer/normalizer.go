package normalizer

import (
	"fmt"
	"go/token"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	. "gopkg.in/bblfsh/sdk.v2/uast/transformer"
	"strings"
	"unicode"
)

var Preprocess = Transformers([][]Transformer{
	{Mappings(
		Map("remove unresolved",
			Part("_", Obj{
				uast.KeyType: String("File"),
				"Unresolved": AnyNode(nil),
			}),
			Part("_", Obj{
				uast.KeyType: String("File"),
			}),
		),
	)},
}...)

var Normalize = Transformers([][]Transformer{
	// The main block of normalization rules.
	{Mappings(Normalizers...)},
}...)

var Normalizers = []Mapping{
	mapUAST("", "Ident", uast.Identifier{},
		map[string]string{
			"NamePos": "start",
		},
		Obj{
			"Name": Var("name"),
		},
		Obj{
			"Name": Var("name"),
		},
	),

	mapUAST("", "BasicLit", uast.String{},
		map[string]string{
			"ValuePos": "start",
		},
		Obj{
			"Kind":  isGoTok(token.STRING),
			"Value": Quote(Var("val")), // TODO: store quote type
		},
		Obj{
			"Value":  Var("val"),
			"Format": String(""),
		},
	),

	mapUAST("", "Comment", uast.Comment{},
		map[string]string{
			"Slash": "start",
		},
		Obj{
			"Text": commentNorm{
				text: "text", block: "block",
				pref: "pref", suff: "suff", tab: "tab",
			},
		},
		Fields{
			{Name: "Block", Op: Var("block")},
			{Name: "Prefix", Op: Var("pref")},
			{Name: "Suffix", Op: Var("suff")},
			{Name: "Tab", Op: Var("tab")},
			{Name: "Text", Op: Var("text")},
		},
	),

	mapUAST("", "BlockStmt", uast.Block{},
		map[string]string{
			"Lbrace": "start",
			"Rbrace": "rbrace", // TODO: off+1 = end
		},
		Obj{
			"List": Var("stmts"),
		},
		Obj{
			"Statements": Var("stmts"),
		},
	),

	mapUAST("uast:Import (all)", "ImportSpec", uast.Import{},
		map[string]string{
			"EndPos": "endp",
		},
		Obj{
			"Comment": Is(nil),
			"Doc":     Is(nil),
			"Name":    Is(nil),
			"Path": uastType(uast.String{}, Obj{
				uast.KeyPos: Var("path_pos"),
				"Value":     Var("path"),
				"Format":    String(""),
			}),
		},
		Obj{
			"Path": uastType(uast.String{}, Obj{
				uast.KeyPos: Var("path_pos"),
				"Value":     Var("path"),
				"Format":    String(""),
			}),
			"All":   Bool(true),
			"Names": Arr(),
		},
	),

	mapUAST("uast:Import (side)", "ImportSpec", uast.Import{},
		map[string]string{
			"EndPos": "endp",
		},
		Obj{
			"Comment": Is(nil),
			"Doc":     Is(nil),
			"Name": uastType(uast.Identifier{}, Obj{
				uast.KeyPos: AnyNode(nil),
				"Name":      String("_"),
			}),
			"Path": uastType(uast.String{}, Obj{
				uast.KeyPos: Var("path_pos"),
				"Value":     Var("path"),
				"Format":    String(""),
			}),
		},
		Obj{
			"Path": uastType(uast.String{}, Obj{
				uast.KeyPos: Var("path_pos"),
				"Value":     Var("path"),
				"Format":    String(""),
			}),
			"All":   Bool(false),
			"Names": Arr(),
		},
	),

	mapUAST("uast:Import (cur)", "ImportSpec", uast.Import{},
		map[string]string{
			"EndPos": "endp",
		},
		Obj{
			"Comment": Is(nil),
			"Doc":     Is(nil),
			"Name": uastType(uast.Identifier{}, Obj{
				uast.KeyPos: AnyNode(nil),
				"Name":      String("."),
			}),
			"Path": uastType(uast.String{}, Obj{
				uast.KeyPos: Var("path_pos"),
				"Value":     Var("path"),
				"Format":    String(""),
			}),
		},
		Obj{
			"Path": uastType(uast.String{}, Obj{
				uast.KeyPos: Var("path_pos"),
				"Value":     Var("path"),
				"Format":    String(""),
			}),
			"All":   Bool(true),
			"Names": Arr(),
			"Scope": String("."), // TODO
		},
	),

	mapUAST("uast:Import (alias)", "ImportSpec", uast.Import{},
		map[string]string{
			"EndPos": "endp",
		},
		Obj{
			"Comment": Is(nil),
			"Doc":     Is(nil),
			"Name": uastType(uast.Identifier{}, Obj{
				uast.KeyPos: Var("alias_pos"),
				"Name":      Var("alias"),
			}),
			"Path": uastType(uast.String{}, Obj{
				uast.KeyPos: Var("path_pos"),
				"Value":     Var("path"),
				"Format":    String(""),
			}),
		},
		Obj{
			"Path": uastType(uast.Alias{}, Obj{
				// FIXME
				//uast.KeyStart: Var("alias_start"),
				//uast.KeyEnd: Var("path_end"),
				"Name": uastType(uast.Identifier{}, Obj{
					uast.KeyPos: Var("alias_pos"),
					"Name":      Var("alias"),
				}),
				"Obj": uastType(uast.String{}, Obj{
					uast.KeyPos: Var("path_pos"),
					"Value":     Var("path"),
					"Format":    String(""),
				}),
			}),
			"All":   Bool(true),
			"Names": Arr(),
		},
	),
	mapUAST("", "FuncDecl", uast.FunctionGroup{},
		nil,
		Obj{
			"Name": Var("name"),
			"Type": Var("type"),
			"Recv": Var("recv"),
			"Body": Var("body"),
			"Doc":  Var("doc"),
		},
		Obj{
			"Nodes": Arr(
				Var("doc"), // FIXME: do not insert if it's nil
				uastType(uast.Alias{}, Obj{
					// FIXME: position
					"Name": Var("name"),
					"Obj": uastType(uast.Function{}, Obj{
						"Type": Var("type"),
						"Body": Var("body"),
						"Recv": Var("recv"), // TODO
					}),
				}),
			),
		},
	),
	mapUAST("", "FuncType", uast.FunctionType{},
		map[string]string{
			"Func": "pos_func",
		},
		Obj{
			"Params": Obj{
				uast.KeyType: String("FieldList"),
				// FIXME: store positions?
				// "Opening" same as start
				// "Closing" same as end
				uast.KeyPos: AnyNode(nil),
				"List":      Var("args"),
			},
			"Results": Opt("results_exists", Var("out")),
		},
		Obj{
			"Args":    Var("args"),
			"Returns": Opt("results_exists", Var("out")),
		},
	),
	mapUAST(" (variadic)", "Field", uast.Argument{},
		map[string]string{
			"Func": "pos_func",
		},
		Obj{
			"Comment": Is(nil), // FIXME: is it possible to attach it?
			"Doc":     Is(nil),
			"Tag":     Is(nil),
			"Names": Arr( // FIXME: name might not exist
				Var("name"),
			),
			"Type": Obj{
				uast.KeyType: String("Ellipsis"),
				// FIXME: store positions?
				// "Ellipsis" same as start
				uast.KeyPos: AnyNode(nil),
				"Elt":       Var("type"),
			},
		},
		Obj{
			"Name":     Var("name"),
			"Type":     Var("type"),
			"Variadic": Bool(true),
		},
	),
}

func uastType(uobj interface{}, op ObjectOp) ObjectOp {
	utyp := uast.TypeOf(uobj)
	if utyp == "" {
		panic(fmt.Errorf("type is not registered: %T", uobj))
	}
	obj := op.Object()
	obj.SetField(uast.KeyType, String(utyp))
	return obj
}

func mapUAST(name, typ string, uobj interface{}, pos map[string]string, src, dst ObjectOp) Mapping {
	utyp := uast.TypeOf(uobj)
	if strings.HasPrefix(name, " ") {
		name = typ + " -> " + utyp + name
	} else if name == "" {
		name = typ + " -> " + utyp
	}
	so, do := src.Object(), dst.Object()

	sp := uastType(uast.Positions{}, Obj{
		uast.KeyStart: Var("start"),
		uast.KeyEnd:   Var("end"),
	}).Object()
	dp := uastType(uast.Positions{}, Obj{
		uast.KeyStart: Var("start"),
		uast.KeyEnd:   Var("end"),
	}).Object()
	for k, v := range pos {
		sp.SetField(k, Var(v))
		if v != "start" && v != "end" {
			dp.SetField(k, Var(v))
		}
	}
	so.SetField(uast.KeyType, String(typ))
	so.SetField(uast.KeyPos, sp)
	do.SetField(uast.KeyPos, dp)
	return Map(name, so, uastType(uobj, do))
}

type commentNorm struct {
	text, block     string
	pref, suff, tab string
}

func (op commentNorm) Check(st *State, n nodes.Node) (bool, error) {
	s, ok := n.(nodes.String)
	if !ok {
		return false, nil
	}
	var (
		text            = string(s)
		pref, suff, tab string
		block           = !strings.HasPrefix(text, "//")
	)
	if !block {
		text = strings.TrimPrefix(text, "//")
	} else {
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
	}

	// find prefix
	i := 0
	for ; i < len(text) && unicode.IsSpace(rune(text[i])); i++ {
	}
	pref = text[:i]
	text = text[i:]

	// find suffix
	i = len(text) - 1
	for ; i >= 0 && unicode.IsSpace(rune(text[i])); i-- {
	}
	suff = text[i+1:]
	text = text[:i+1]

	// TODO: set tab

	err := st.SetVars(Vars{
		op.text:  nodes.String(text),
		op.pref:  nodes.String(pref),
		op.suff:  nodes.String(suff),
		op.tab:   nodes.String(tab),
		op.block: nodes.Bool(block),
	})
	return err == nil, err
}

func (op commentNorm) Construct(st *State, n nodes.Node) (nodes.Node, error) {
	var (
		text, pref, suff, tab nodes.String
		block                 nodes.Bool
	)

	err := st.MustGetVars(VarsPtrs{
		op.text: &text, op.block: &block,
		op.pref: &pref, op.suff: &suff, op.tab: &tab,
	})
	if err != nil {
		return nil, err
	}
	// FIXME: handle tab
	text = pref + text + suff
	if !block {
		return "//" + text, nil
	}
	return "/*" + text + "*/", nil
}
