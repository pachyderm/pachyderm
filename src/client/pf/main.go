package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
)

func do() {
	ctx := &config.Context{}
	pf, err := client.NewPortForwarder(ctx, "default")
	if err != nil {
		log.Fatal(err)
	}
	port, err := pf.Run("nginx", 1234, 80)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(port)
	res, err := http.Get("http://localhost:1234")
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := httputil.DumpResponse(res, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", bytes)
	time.Sleep(10 * time.Minute)
	pf.Close()
}

func main() {
	for i := 0; i < 1; i++ {
		do()
	}
}
