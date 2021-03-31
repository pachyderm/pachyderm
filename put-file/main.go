package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"google.golang.org/grpc"
)

var (
	addr      = flag.String("addr", "dns:///172.18.0.2:30650", "address of grpc server")
	nSenders  = flag.Int("senders", 1, "number of sending goroutines")
	nMessages = flag.Int("messages", 1, "number of messages to send per goroutine")
	msgLen    = flag.Int("length", 1e3, "size of each file")
)

type stat struct {
	N                           int
	SendDuration, CloseDuration time.Duration
}

func main() {
	flag.Parse()
	cc, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	var bytes int
	c := pfs.NewAPIClient(cc)
	doneCh := make(chan stat)

	start := time.Now()
	for i := 0; i < *nSenders; i++ {
		go func(i int) {
			stats := stat{}
			ctx := context.Background()
			buf := make([]byte, *msgLen)

			pf, err := c.PutFile(ctx)
			if err != nil {
				log.Fatal(err)
			}
			sendStart := time.Now()
			for j := 0; j < *nMessages; j++ {
				path := fmt.Sprintf("data-%d", i**nMessages+j)
				if err := pf.Send(&pfs.PutFileRequest{
					File: &pfs.File{
						Commit: &pfs.Commit{
							Repo: &pfs.Repo{
								Name: "data",
							},
							ID: "master",
						},
						Path: path,
					},
					Value: buf,
				}); err != nil {
					log.Fatal(err)
				}
				stats.N += *msgLen
			}
			stats.SendDuration = time.Since(sendStart)
			closeStart := time.Now()
			if _, err := pf.CloseAndRecv(); err != nil {
				log.Fatal(err)
			}
			stats.CloseDuration = time.Since(closeStart)
			doneCh <- stats
		}(i)
	}
	for i := 0; i < *nSenders; i++ {
		stat := <-doneCh
		bytes += stat.N
		fmt.Printf("%#v\n", stat)
		dur := time.Since(start)
		fmt.Printf("%d threads done; %s bytes sent in %s -> %s/s\n", i+1, humanize.Bytes(uint64(bytes)), dur, humanize.Bytes(uint64(float64(bytes)/dur.Seconds())))
	}
	close(doneCh)
	cc.Close()
}
