package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	reg       *regexp.Regexp
	inputDir  string
	outputDir string
)

func sanitize(word string) []string {
	sanitized := reg.ReplaceAllString(word, " ")
	return strings.Split(strings.ToLower(sanitized), " ")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatalf("expect two arguments; got %v", len(args))
	}

	var err error
	reg, err = regexp.Compile(`[^A-Za-z]+`)
	if err != nil {
		log.Fatal(err)
	}

	inputDir = args[0]
	outputDir = args[1]

	wordMap := make(map[string]int)
	if err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		log.Printf("scanning %v", path)
		f, err := os.Open(path)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanWords)
		count := 0
		for scanner.Scan() {
			count += 1
			for _, word := range sanitize(scanner.Text()) {
				if word != "" {
					wordMap[word] = wordMap[word] + 1
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		log.Printf("found %d words in %s", count, path)

		if err := f.Close(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	for word, count := range wordMap {
		if err := ioutil.WriteFile(filepath.Join(outputDir, word), []byte(strconv.Itoa(count)+"\n"), 0644); err != nil {
			log.Fatal(err)
		}
	}
}
