package convert

import . "go/token"

var tokens = map[string]Token{
	"ILLEGAL":     ILLEGAL,
	"EOF":         EOF,
	"COMMENT":     COMMENT,
	"IDENT":       IDENT,
	"INT":         INT,
	"FLOAT":       FLOAT,
	"IMAG":        IMAG,
	"CHAR":        CHAR,
	"STRING":      STRING,
	"+":           ADD,
	"-":           SUB,
	"*":           MUL,
	"/":           QUO,
	"%":           REM,
	"&":           AND,
	"|":           OR,
	"^":           XOR,
	"<<":          SHL,
	">>":          SHR,
	"&^":          AND_NOT,
	"+=":          ADD_ASSIGN,
	"-=":          SUB_ASSIGN,
	"*=":          MUL_ASSIGN,
	"/=":          QUO_ASSIGN,
	"%=":          REM_ASSIGN,
	"&=":          AND_ASSIGN,
	"|=":          OR_ASSIGN,
	"^=":          XOR_ASSIGN,
	"<<=":         SHL_ASSIGN,
	">>=":         SHR_ASSIGN,
	"&^=":         AND_NOT_ASSIGN,
	"&&":          LAND,
	"||":          LOR,
	"<-":          ARROW,
	"++":          INC,
	"--":          DEC,
	"==":          EQL,
	"<":           LSS,
	">":           GTR,
	"=":           ASSIGN,
	"!":           NOT,
	"!=":          NEQ,
	"<=":          LEQ,
	">=":          GEQ,
	":=":          DEFINE,
	"...":         ELLIPSIS,
	"(":           LPAREN,
	"[":           LBRACK,
	"{":           LBRACE,
	",":           COMMA,
	".":           PERIOD,
	")":           RPAREN,
	"]":           RBRACK,
	"}":           RBRACE,
	";":           SEMICOLON,
	":":           COLON,
	"break":       BREAK,
	"case":        CASE,
	"chan":        CHAN,
	"const":       CONST,
	"continue":    CONTINUE,
	"default":     DEFAULT,
	"defer":       DEFER,
	"else":        ELSE,
	"fallthrough": FALLTHROUGH,
	"for":         FOR,
	"func":        FUNC,
	"go":          GO,
	"goto":        GOTO,
	"if":          IF,
	"import":      IMPORT,
	"interface":   INTERFACE,
	"map":         MAP,
	"package":     PACKAGE,
	"range":       RANGE,
	"return":      RETURN,
	"select":      SELECT,
	"struct":      STRUCT,
	"switch":      SWITCH,
	"type":        TYPE,
	"var":         VAR,
}
