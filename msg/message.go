package msg

import "go/ast"

const (
	// Ok status code.
	Ok = "ok"
	// Error status code.
	Error = "error"
	// Fatal status code.
	Fatal = "fatal"
	// ParseAst is the Action identifier to parse an AST.
	ParseAst = "ParseAST"
)

// Request is the message the driver receives. It marshals to Messagepack.
type Request struct {
	Action          string `codec:"action"`
	Language        string `codec:"language,omitempty"`
	LanguageVersion string `codec:"language_version,omitempty"`
	Content         string `codec:"content"`
}

// Response is the message replies. It marshals to Messagepack.
type Response struct {
	Status          string    `codec:"status"`
	Errors          []string  `codec:"errors,omitempty"`
	Driver          string    `codec:"driver"`
	Language        string    `codec:"language"`
	LanguageVersion string    `codec:"language_version"`
	AST             *ast.File `codec:"ast"`
}
