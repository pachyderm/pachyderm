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

var logger = log.Default()

func main() {
	logger.Println("Running merger code")
	inputFolder := os.Args[1]
	outputFolder := os.Args[2]

	mergeArgs := []string{"tool", "covdata", "merge", "-pcombine", "-i"}
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
			out, _ := exec.Command("ls", canonPath).CombinedOutput() // DNJ TODO - delete
			logger.Printf("%s folder ls: %s \n", canonPath, string(out))
		}
		return nil
	})
	fmt.Println("files: ", files)
	branch := strings.Split( // DNJ TODO - extract/share this among all pipelines?
		strings.TrimLeft(files[0], string(filepath.Separator)),
		string(filepath.Separator),
	)[2] // /pfs/repo-name/branch

	// merge coverage
	mergeOutputFolder := filepath.Join(outputFolder, branch, "raw")
	err = os.MkdirAll(mergeOutputFolder, fs.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}

	mergeArgs = append(mergeArgs, strings.Join(files, ","), "-o", mergeOutputFolder)
	_, err = runGoCommand(mergeArgs)
	if err != nil {
		log.Fatal(err)
	}

	// format coverage as text file
	err = os.MkdirAll(filepath.Join(outputFolder, branch), fs.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}
	out, _ := exec.Command("ls", mergeOutputFolder).CombinedOutput() // DNJ TODO - delete
	logger.Printf("%s folder ls: %s \n", mergeOutputFolder, string(out))
	formatArgs = append(formatArgs, mergeOutputFolder, "-o", filepath.Join(outputFolder, branch, "profile.txt"))
	_, err = runGoCommand(formatArgs)
	if err != nil {
		log.Fatal(err)
	}
}

func runGoCommand(args []string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	fmt.Println("Running command: ", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("Command failed: %v \n", string(out))
		return nil, err
	} else {
		logger.Printf("Command succeeded: %v \n", string(out))
	}
	return out, nil
}
