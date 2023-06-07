package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/open-policy-agent/opa/rego"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: test-expr SPEC VALIDATOR")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	query, err := rego.New(
		rego.Query("not data.hpe.ai.pachyderm.examples.validation.deny"),
		rego.Module("example.rego", string(b)),
	).PrepareForEval(ctx)

	if err != nil {
		panic(err)
	}

	f, err = os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	b, err = io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var jsonObject any
	if err := json.Unmarshal(b, &jsonObject); err != nil {
		panic(err)
	}

	results, err := query.Eval(ctx, rego.EvalInput(jsonObject))
	if err != nil {
		panic(err)
	}
	if !results.Allowed() {
		fmt.Fprintln(os.Stderr, "not allowed")
		os.Exit(1)
	}
}
