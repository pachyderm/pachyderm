package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.starlark.net/starlark"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: test-expr SPEC VALIDATOR")
		os.Exit(1)
	}
	var t = &starlark.Thread{Name: "default"}
	var errors []string
	e := starlark.NewBuiltin("error", func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		errors = append(errors, args.Index(0).String())
		return starlark.None, nil
	})
	g, err := starlark.ExecFile(t, os.Args[2], nil, starlark.StringDict{"error": e})
	if err != nil {
		panic(err)
	}
	if len(errors) > 0 {
		for _, errorMessage := range errors {
			fmt.Fprintln(os.Stderr, errorMessage)
		}
		os.Exit(0)
	}
	validate := g["validate"]
	var request = new(pps.CreatePipelineRequest)
	b, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	var jsonObject any
	if err := json.Unmarshal(b, &jsonObject); err != nil {
		panic(err)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	b, err = io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, &jsonObject); err != nil {
		panic(err)
	}
	v, err := jsonToValue(jsonObject)
	if err != nil {
		panic(err)
	}
	v, err = starlark.Call(t, validate, starlark.Tuple{v}, nil)
	if err != nil {
		panic(err)
	}
	if len(errors) > 0 {
		for _, errorMessage := range errors {
			fmt.Fprintln(os.Stderr, errorMessage)
		}
		os.Exit(0)
	}
}

func jsonToValue(j any) (starlark.Value, error) {
	switch j := j.(type) {
	case string:
		return starlark.String(j), nil
	case map[string]any:
		m := starlark.NewDict(len(j))
		for k, v := range j {
			jv, err := jsonToValue(v)
			if err != nil {
				return nil, fmt.Errorf("could not convert field %q: %w", k, err)
			}
			m.SetKey(starlark.String(k), jv)
		}
		return m, nil
	case []any:
		var ll = make([]starlark.Value, len(j))
		for i, v := range j {
			var err error
			if ll[i], err = jsonToValue(v); err != nil {
				return nil, fmt.Errorf("could not convert item %d: %w", i, err)
			}
		}
		return starlark.NewList(ll), nil
	case bool:
		return starlark.Bool(j), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to starlark value", j)
	}
}
