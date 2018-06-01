package normalizer

import (
	"go/token"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	. "gopkg.in/bblfsh/sdk.v2/uast/transformer"
	"strings"
	"unicode"
)

var Preprocess = Transformers([][]Transformer{
	{Mappings(
		MapPart("_", MapObj( // remove unresolved
			Obj{
				uast.KeyType: String("File"),
				"Unresolved": AnyNode(nil),
			},
			Obj{
				uast.KeyType: String("File"),
			},
		)),
	)},
}...)

var Normalize = Transformers([][]Transformer{
	// The main block of normalization rules.
	{Mappings(Normalizers...)},
}...)

var Normalizers = []Mapping{
	MapSemanticPos("Ident", uast.Identifier{},
		map[string]string{
			"NamePos": "start",
		},
		ObjMap{"Name": Var("name")},
	),

	MapSemanticPos("BasicLit", uast.String{},
		map[string]string{
			"ValuePos": "start",
		},
		MapObj(
			Obj{
				"Kind":  isGoTok(token.STRING),
				"Value": Quote(Var("val")), // TODO: store quote type
			},
			Obj{
				"Value":  Var("val"),
				"Format": String(""),
			},
		),
	),

	MapSemanticPos("Comment", uast.Comment{},
		map[string]string{
			"Slash": "start",
		},
		MapObj(
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
	),

	MapSemanticPos("BlockStmt", uast.Block{},
		map[string]string{
			"Lbrace": "start",
			"Rbrace": "rbrace", // TODO: off+1 = end
		},
		MapObj(
			Obj{
				"List": Var("stmts"),
			},
			Obj{
				"Statements": Var("stmts"),
			},
		),
	),
	// all-in-one
	MapSemanticPos("ImportSpec", uast.Import{},
		map[string]string{
			"EndPos": "endp",
		},
		MapObj(
			Obj{
				"Comment": Is(nil),
				"Doc":     Is(nil),
				"Path":    Var("path"),
				"Name": Cases("case",
					// case 1: no alias for the import
					Is(nil),
					// case 2: side-effect import
					UASTType(uast.Identifier{}, Obj{
						uast.KeyPos: AnyNode(nil),
						"Name":      String("_"),
					}),
					// case 3: import to the package scope
					UASTType(uast.Identifier{}, Obj{
						uast.KeyPos: AnyNode(nil),
						"Name":      String("."),
					}),
				),
			},
			// ->
			CasesObj("case",
				// common
				Obj{
					"Path": Var("path"),
				},
				Objs{
					// case 1: no alias for the import
					{
						"All":    Bool(true),
						"Names":  Arr(),
						"Target": Is(nil), // TODO
					},
					// case 2: side-effect import
					{
						"All":    Bool(false),
						"Names":  Arr(),
						"Target": Is(nil), // TODO
					},
					// case 3: import to the package scope
					{
						"All":    Bool(true),
						"Names":  Arr(),
						"Target": Obj{"Scope": String(".")}, // TODO
					},
				},
			),
		),
	),

	// alias
	MapSemanticPos("ImportSpec", uast.Import{},
		map[string]string{
			"EndPos": "endp",
		},

		MapObj(
			Obj{
				"Comment": Is(nil),
				"Doc":     Is(nil),
				"Name": UASTType(uast.Identifier{}, Obj{
					uast.KeyPos: Var("alias_pos"),
					"Name":      Var("alias"),
				}),
				"Path": UASTType(uast.String{}, Obj{
					uast.KeyPos: Var("path_pos"),
					"Value":     Var("path"),
					"Format":    String(""),
				}),
			},
			// ->
			Obj{
				"Path": UASTType(uast.Alias{}, Obj{
					// FIXME
					//uast.KeyStart: Var("alias_start"),
					//uast.KeyEnd: Var("path_end"),
					"Name": UASTType(uast.Identifier{}, Obj{
						uast.KeyPos: Var("alias_pos"),
						"Name":      Var("alias"),
					}),
					"Node": UASTType(uast.String{}, Obj{
						uast.KeyPos: Var("path_pos"),
						"Value":     Var("path"),
						"Format":    String(""),
					}),
				}),
				"All":   Bool(true),
				"Names": Arr(),
			},
		),
	),
	// FIXME: it affects struct fields as well
	MapSemantic("Field", uast.Argument{},
		MapObj(
			Obj{
				"Comment": Is(nil), // FIXME: is it possible to attach it?
				"Doc":     Is(nil),
				"Tag":     Is(nil),
				"Names": Cases("names",
					Is(nil),
					Arr( // FIXME: there might be multiple names
						Var("name"),
					),
				),
				"Type": Cases("variadic",
					// case 1: variadic
					Obj{
						uast.KeyType: String("Ellipsis"),
						// FIXME: store positions?
						// "Ellipsis" same as start
						uast.KeyPos: AnyNode(nil),
						"Elt":       Var("type"),
					},
					// case 2: normal arg
					Check(
						Not(Has{uast.KeyType: String("Ellipsis")}),
						Var("type"),
					),
				),
			},
			CasesObj("variadic",
				// common
				Obj{
					"Name": Cases("names",
						Is(nil),
						Check(NotNil(), Var("name")),
					),
					"Type":     Var("type"),
					"Receiver": Bool(false),
				},
				Objs{
					// case 1: variadic
					{"Variadic": Bool(true)},
					// case 2: normal arg
					{"Variadic": Bool(false)},
				},
			),
		),
	),
	MapSemanticPos("FuncType", uast.FunctionType{},
		map[string]string{
			"Func": "pos_func",
		},
		MapObj(
			Obj{
				"Params": Obj{
					uast.KeyType: String("FieldList"),
					// FIXME: store positions?
					// "Opening" same as start
					// "Closing" same as end
					uast.KeyPos: AnyNode(nil),
					"List": Cases("list",
						Is(nil),
						Check(NotNil(), Var("args")),
					),
				},
				"Results": Cases("res",
					Is(nil),
					Obj{
						uast.KeyType: String("FieldList"),
						// FIXME: store positions?
						// "Opening" same as start
						// "Closing" same as end
						uast.KeyPos: AnyNode(nil),
						"List":      Var("out"),
					},
				),
			},
			Obj{
				"Arguments": Cases("list",
					Arr(),
					Check(NotNil(), Var("args")),
				),
				"Returns": Cases("res",
					Is(nil),
					Var("out"),
				),
			},
		),
	),
	MapSemantic("FuncDecl", uast.FunctionGroup{},
		MapObj(
			CasesObj("recv_case",
				// common
				Obj{
					"Name": Var("name"),
					"Body": Var("body"),
					"Doc":  Var("doc"),
				},
				Objs{
					// case 1: no receiver
					{
						"Type": Var("type"),
						"Recv": Is(nil),
					},
					// case 2: receiver
					{
						"Type": UASTTypePart("type", uast.FunctionType{}, Obj{
							"Arguments": Var("args"),
						}),
						"Recv": Obj{
							uast.KeyType: String("FieldList"),
							// FIXME: store positions?
							// "Opening" same as start
							// "Closing" same as end
							uast.KeyPos: AnyNode(nil),
							"List": One(
								UASTTypePart("recv", uast.Argument{}, Obj{
									"Receiver": Bool(false),
								}),
							),
						},
					},
				},
			),
			// ->
			Obj{
				"Nodes": Arr(
					Var("doc"), // FIXME: do not insert if it's nil
					UASTType(uast.Alias{}, Obj{
						// FIXME: position
						"Name": Var("name"),
						"Node": UASTType(uast.Function{}, CasesObj("recv_case",
							// common
							Obj{
								"Body": Var("body"),
							},
							Objs{
								// case 1: no receiver - store type as-is
								{
									"Type": Var("type"),
								},
								// case 2: receiver - need to inject as a first argument with a flag
								{
									"Type": UASTTypePart("type", uast.FunctionType{}, Obj{
										"Arguments": PrependOne(
											UASTTypePart("recv", uast.Argument{}, Obj{
												"Receiver": Bool(true),
											}),
											Var("args"),
										),
									}),
								},
							},
						)),
					}),
				),
			},
		),
	),
}

type commentNorm struct {
	text, block     string
	pref, suff, tab string
}

func (commentNorm) Kinds() nodes.Kind {
	return nodes.KindString
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
