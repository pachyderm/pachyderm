package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"

	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: test-expr SPEC VALIDATOR")
		os.Exit(1)
	}
	var request = new(pps.CreatePipelineRequest)
	b, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	var jsonObject any
	if err := json.Unmarshal(b, &jsonObject); err != nil {
		panic(err)
	}
	var errors []string
	env := map[string]interface{}{
		"sprint": fmt.Sprint,
		"error":  func(msg string) string { errors = append(errors, msg); return msg },
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
	env["request"] = jsonObject

	if f, err = os.Open(os.Args[2]); err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	var rules []*vm.Program
	for s.Scan() {
		p, err := expr.Compile(s.Text())
		if err != nil {
			panic(err)
		}
		rules = append(rules, p)
	}

	for _, rule := range rules {
		output, err := expr.Run(rule, env)
		if err != nil {
			panic(err)
		}
		switch output := output.(type) {
		case bool:
			if !output {
				fmt.Fprintln(os.Stderr, "rule failed")
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "unexpected rule result type (%T): %v\n", output, output)
			os.Exit(1)
		}
	}
	if len(errors) > 0 {
		for _, errorMessage := range errors {
			fmt.Fprintln(os.Stderr, errorMessage)
		}
		os.Exit(1)
	}
}
