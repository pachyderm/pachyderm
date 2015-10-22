[![CircleCI](https://circleci.com/gh/peter-edge/go-env/tree/master.png)](https://circleci.com/gh/peter-edge/go-env/tree/master)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/env)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/go-env/blob/master/LICENSE)

Package env handles environment variables in a structured manner. It uses reflection
to set fields on a struct pointer passed to the main public method, `Populate`.

Example (from the test):

```go
type testEnv struct {
	RequiredString string `env:"REQUIRED_STRING,required"`
	OptionalString string `env:"OPTIONAL_STRING"`
	OptionalInt    int    `env:"OPTIONAL_INT"`
	OptionalBool   bool   `env:"OPTIONAL_BOOL"`
	OptionalStruct struct {
		StructOptionalInt int `env:"STRUCT_OPTIONAL_INT"`
	} `env:"OPTIONAL_STRUCT,match=FOO"`
	OptionalStructTwo struct {
		StructOptionalInt int `env:"STRUCT_OPTIONAL_INT"`
	} `env:"OPTIONAL_STRUCT_TWO,match=^(FOO|BAR)$"`
}
```

To populate this, one would call:

```go
testEnv := &testEnv{}
if err := env.Populate(testEnv, PopulateOptionsP{}); err != nil {
	return err
}
```

Non-struct fields have the name of the environment variable, optionally followed
by a "required" option, which will throw an error if the environment variable is
not set. Struct fields have the name of an environment variable, (not) optionally
followed at a "match=REGEX" option, which says that if the value of the environment
variable matches the regex, set the fields within the struct. This is useful for
situations such as:

```go
type env struct {
	QueueType string `env:"QUEUE_TYPE,required"`
	Sqs struct {
		AwsRegion string `env:"AWS_REGION,required"`
	} `env:"QUEUE_TYPE,match=^SQS$"`
	Redis struct {
		DatabaseURL string `env:"DATABASE_URL,required"`
	} `env:"QUEUE_TYPE,match=^REDIS$"`
}
```

`PopulateOptions` has two fields:

* `RestrictTo`: this is a list of all allowed environment variables on structs. This is
useful if you have a bunch of mini-services within a repo, and you want to record all
variables that are used across your services in one comment place in code.
* `Decoders`: you can set additional Decoders (both an env file decoder and a JSON decoder
are provided) to provider additional environment variables on top of those set at the system
level. Decoders overwrite system level values, and Decoders later in the slice override those
earlier in the slice.
