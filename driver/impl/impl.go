package impl

import (
	"github.com/bblfsh/go-driver/driver/golang"
	"github.com/bblfsh/sdk/v3/driver/server"
)

func init() {
	server.DefaultDriver = golang.NewDriver()
}
