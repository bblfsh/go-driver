package main

import (
	"github.com/bblfsh/go-driver/driver/golang"
	"gopkg.in/bblfsh/sdk.v2/driver/native"
)

func main() {
	native.Main(golang.NewDriver())
}
