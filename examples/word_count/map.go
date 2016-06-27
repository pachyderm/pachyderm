package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strconv"
	"strings"
)

var (
	reg *regexp.Regexp
)

func sanitize(word string) []string {
	sanitized := reg.ReplaceAllString(word, " ")
	return strings.Split(strings.ToLower(sanitized), " ")
}

func shuffle(slice []os.FileInfo) {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatalf("expect two arguemnts; got %v", len(args))
	}

	var err error
	reg, err = regexp.Compile(`[^A-Za-z]+`)
	if err != nil {
		log.Fatal(err)
	}

	inputDir := args[0]
	outputDir := args[1]

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		log.Fatal(err)
	}

	// we want to process files in a random order in order to
	// avoid flash-crowd behavior
	shuffle(files)

	wordMap := make(map[string]int)

	for _, file := range files {
		log.Printf("scanning file %v", file.Name())
		f, err := os.Open(filepath.Join(inputDir, file.Name()))
		if err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		}

		log.Printf("found %d words in %s", count, file.Name())

		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}

	var i int
	for word, count := range wordMap {
		i++
		if err := ioutil.WriteFile(filepath.Join(outputDir, word), []byte(strconv.Itoa(count)+"\n"), 0644); err != nil {
			log.Fatal(err)
		}
		if i%1000 == 0 {
			memprofile, err := os.Create(fmt.Sprintf("/tmp/memprofile-%d", i/1000))
			if err != nil {
				panic(err)
			}
			pprof.WriteHeapProfile(memprofile)
			memprofile.Close()
		}
	}
}
