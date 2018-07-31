package main

import (
	_ "github.com/bblfsh/go-driver/driver/impl"
	"github.com/bblfsh/go-driver/driver/normalizer"

	"gopkg.in/bblfsh/sdk.v2/driver/server"
)

func main() {
	server.Run(normalizer.Transforms)
}
