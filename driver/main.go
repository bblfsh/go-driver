package main

import (
	_ "github.com/bblfsh/go-driver/driver/impl"
	"github.com/bblfsh/go-driver/driver/normalizer"

	"gopkg.in/bblfsh/sdk.v2/sdk/driver"
)

func main() {
	driver.Run(driver.Transforms{
		Preprocess: normalizer.Preprocess,
		Normalize:  normalizer.Normalize,
		Native:     normalizer.Native,
		Code:       normalizer.Code,
	})
}
