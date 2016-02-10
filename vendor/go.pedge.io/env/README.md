[![CircleCI](https://circleci.com/gh/peter-edge/env-go/tree/master.png)](https://circleci.com/gh/peter-edge/env-go/tree/master)
[![Go Report Card](http://goreportcard.com/badge/peter-edge/env-go)](http://goreportcard.com/report/peter-edge/env-go)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/env)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/env-go/blob/master/LICENSE)

Package env handles environment variables in a structured manner. It uses reflection
to set fields on a struct pointer passed to the main public method, `Populate`.

Example (from the test):

```go
type subEnv struct {
	SubEnvRequiredString        string `env:"SUB_ENV_REQUIRED_STRING,required"`
	SubEnvOptionalString        string `env:"SUB_ENV_OPTIONAL_STRING"`
	SubEnvOptionalUint16Default uint16 `env:"SUB_ENV_OPTIONAL_UINT16_DEFAULT,default=1024"`
}

type testEnv struct {
	RequiredString string `env:"REQUIRED_STRING,required"`
	OptionalString string `env:"OPTIONAL_STRING"`
	OptionalInt    int    `env:"OPTIONAL_INT"`
	OptionalBool   bool   `env:"OPTIONAL_BOOL"`
	SubEnv         subEnv
	Struct         struct {
		StructOptionalInt int `env:"STRUCT_OPTIONAL_INT"`
	}
	OptionalStringDefault string `env:"OPTIONAL_STRING_DEFAULT,default=foo"`
}
```

To populate this, one would call:

```go
testEnv := &testEnv{}
if err := env.Populate(testEnv); err != nil {
	return err
}
```
