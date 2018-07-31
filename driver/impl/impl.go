package impl

import (
	"github.com/bblfsh/go-driver/driver/golang"
	"gopkg.in/bblfsh/sdk.v2/driver/server"
)

func init() {
	server.DefaultDriver = golang.NewDriver()
}
