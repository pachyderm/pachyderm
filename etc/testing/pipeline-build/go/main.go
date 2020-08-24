package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/keltia/leftpad"
)

const (
	padLength       = 4
	inputDirectory  = "/pfs/in"
	outputDirectory = "/pfs/out"
)

func main() {
	padChar := '0'
	if len(os.Args) > 1 {
		padChar = rune(os.Args[1][0])
	}

	fis, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		panic(err)
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		inPath := path.Join(inputDirectory, fi.Name())
		inBytes, err := ioutil.ReadFile(inPath)
		if err != nil {
			panic(err)
		}
		in := string(inBytes)

		outPath := path.Join(outputDirectory, fi.Name())
		out, err := leftpad.PadChar(in, padLength, padChar)
		if err != nil {
			panic(err)
		}
		if err = ioutil.WriteFile(outPath, []byte(out), os.ModePerm); err != nil {
			panic(err)
		}
	}
}
