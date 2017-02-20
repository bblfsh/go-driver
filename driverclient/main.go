package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/src-d/lang-parsers/go/go-driver/msg"

	"github.com/jessevdk/go-flags"
	"github.com/ugorji/go/codec"
)

type options struct {
	File            string `short:"f" long:"file" description:"Source code file" required:"true"`
	Language        string `short:"l" long:"language" description:"File's source code language" default:""`
	LanguageVersion string `short:"v" long:"version" description:"File's source code language version" default:""`
}

func main() {
	var opt options
	parser := flags.NewParser(&opt, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	f, err := os.Open(opt.File)
	if err != nil {
		log.Fatal(err)
	}

	source, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	var handle codec.MsgpackHandle
	mpEnc := codec.NewEncoder(os.Stdout, &handle)
	mpEnc.MustEncode(&msg.Request{
		Action:          msg.ParseAst,
		Language:        opt.Language,
		LanguageVersion: opt.LanguageVersion,
		Content:         string(source),
	})
}
