package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	logger := log.Default()
	logger.Println("Running merger code")
	inputFolder := os.Args[1]
	outputFolder := os.Args[2]

	formatArgs := []string{"tool", "covdata", "textfmt", "-i"}

	sym, err := filepath.EvalSymlinks(inputFolder)
	if err != nil {
		logger.Fatal(err)
	}
	files := []string{}
	filepath.WalkDir(sym, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() { // DNJ TODO - check for non-code coverage files
			canonPath := strings.Replace(path, sym, inputFolder, 1)
			if canonPath != inputFolder {
				files = append(files, strings.Replace(path, sym, inputFolder, 1))
			}
		}
		return nil
	})
	fmt.Println("files: ", files)
	branch := strings.Split( // DNJ TODO - extract/share this among all pipelines?
		strings.TrimLeft(files[0], string(filepath.Separator)),
		string(filepath.Separator),
	)[2] // /pfs/repo-name/branch
	err = os.MkdirAll(filepath.Join(outputFolder, branch), fs.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}
	formatArgs = append(formatArgs, strings.Join(files, ","), "-o", filepath.Join(outputFolder, branch, "profile.txt"))
	// DNJ TODO - do we need explicit merge call
	cmd := exec.Command("go", formatArgs...)
	fmt.Println("Running command: ", cmd.String())
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Printf("Command failed: %v \n", string(out))
		log.Fatal(err)
	} else {
		logger.Printf("Command succeeded: %v \n", string(out))
	}
	// check for consumed file right before save
	// put cov file in codecov - deferred pipeline pushes to codecov
}
