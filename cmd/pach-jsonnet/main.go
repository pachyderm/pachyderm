package main

import (
	"fmt"

	"github.com/google/go-jsonnet"
)

func main() {
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{JPaths: []string{"."}})
	json, err := vm.EvaluateFile("effective.jsonnet")
	if err != nil {
		panic(err)
	}
	fmt.Println(json)
}
