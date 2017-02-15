package msg

import "go/ast"

const (
	// Ok status code.
	Ok = "ok"
	// Error status code. It is replied when the driver has got the AST with errors.
	Error = "error"
	// Fatal status code. It is replied when the driver hasn't could get the AST.
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

// Response is the replied message. It marshals to Messagepack.
type Response struct {
	Status          string    `codec:"status"`
	Errors          []string  `codec:"errors,omitempty"`
	Driver          string    `codec:"driver"`
	Language        string    `codec:"language"`
	LanguageVersion string    `codec:"language_version"`
	AST             *ast.File `codec:"ast"`
}
