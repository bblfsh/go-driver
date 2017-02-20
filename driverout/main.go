package main

import (
	"fmt"
	"os"

	"github.com/src-d/babelfish-go-driver/msg"
	"github.com/ugorji/go/codec"
)

func main() {
	var handle codec.MsgpackHandle
	mpDec := codec.NewDecoder(os.Stdin, &handle)

	var t interface{}
	mpDec.MustDecode(&t)
	m := t.(map[interface{}]interface{})

	fmt.Println("")
	fmt.Printf("%#v\n", m)
	fmt.Println("")

	res := &msg.Response{}
	for k, v := range m {
		switch {
		case k == "status":
			res.Status = string(v.([]uint8))
		case k == "errors":
			res.Errors = v.([]string)
		case k == "language":
			res.Language = string(v.([]uint8))
		case k == "language_version":
			res.LanguageVersion = string(v.([]uint8))
		case k == "driver":
			res.Driver = string(v.([]uint8))
		case k == "ast":
			res.AST = nil
		}
	}

	fmt.Printf("%#v\n", res)
}
