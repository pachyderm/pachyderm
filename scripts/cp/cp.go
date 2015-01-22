package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

func main() {
	log.SetFlags(log.Lshortfile)
	files := make(chan string, 1000)
	defer close(files)

	// spawn four worker goroutines
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range files {
				file, err := os.Open(name)
				if err != nil {
					log.Print(err)
				}
				defer file.Close()

				resp, err := http.Post("http://"+os.Args[2]+"/file/"+name, "application/text", file)
				if err != nil {
					log.Print(err)
				}
				resp.Body.Close()
			}
		}()
	}

	dir, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	counter := 0
	var names []string
	for names, err = dir.Readdirnames(1000); err == nil; names, err = dir.Readdirnames(1000) {
		for _, name := range names {
			files <- path.Join(os.Args[1], name)
		}
		counter += len(names)
		log.Printf("Sent %d so far.", counter)
	}
	if err != io.EOF {
		log.Fatal(err)
	}

}
