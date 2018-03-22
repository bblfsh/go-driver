package main

import (
	"github.com/bblfsh/go-driver/driver/normalizer"

	"gopkg.in/bblfsh/sdk.v1/sdk/driver"
)

func main() {
	driver.Run(driver.Transforms{
		Native: normalizer.Native,
		Code:   normalizer.Code,
	})
}
