package main

import (
	"github.com/bblfsh/go-driver/driver/golang"
	"github.com/bblfsh/sdk/v3/driver/native"
)

func main() {
	native.Main(golang.NewDriver())
}
