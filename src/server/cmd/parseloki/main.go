package main

import (
	"bufio"
	"log"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/server"
)

func main() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		var result pps.LogMessage
		if err := server.ParseLokiLine(s.Text(), &result); err != nil {
			log.Fatal(err)
		}
		log.Println(result)
	}
}
