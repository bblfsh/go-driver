package main

import (
	"log"
	"os"

	"github.com/src-d/lang-parsers/go/go-driver/msg"

	"flag"
	"io/ioutil"

	"github.com/ugorji/go/codec"
)

func main() {
	var file string
	flag.StringVar(&file, "f", "", "file to get the AST")
	flag.Parse()

	if file == "" {
		log.Fatal("file didn't pass as an argument")
	}

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	source, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	var mpHandle codec.MsgpackHandle
	mpEnc := codec.NewEncoder(os.Stdout, &mpHandle)
	mpEnc.MustEncode(&msg.Request{
		Action:  msg.ParseAst,
		Content: string(source),
	})
}
