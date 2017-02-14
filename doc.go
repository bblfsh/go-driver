/*
babelfish-go-driver is a process that reads a msg.Request, in Messagepack format, from standard input.

Then, it takes the "content"(a go source code string) and extracts the AST from this code. Go-driver creates
a msg.Response wich contains the AST, and serliazes it to Messagepack.

Eventually, the process writes the response to standard output.
*/
package main
