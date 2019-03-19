package normalizer

import "gopkg.in/bblfsh/sdk.v2/driver"

var Transforms = driver.Transforms{
	Namespace:      "go",
	Preprocess:     Preprocess,
	PreprocessCode: PreprocessCode,
	Normalize:      Normalize,
	Annotations:    Native,
	Code:           Code,
}
