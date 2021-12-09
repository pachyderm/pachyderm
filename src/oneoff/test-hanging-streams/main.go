package main

import (
	"context"
	"flag"
	"log"
	"sync/atomic"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	nstreams = flag.Int("n", 10, "number of streams to hold open")
)

func main() {
	flag.Parse()

	cc, err := grpc.Dial("localhost:80", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	doneCh := make(chan int)
	ppsC := pps.NewAPIClient(cc)
	var open, errs, gotMsg int64
	for i := 0; i < *nstreams; i++ {
		go func(i int) {
			defer func() {
				doneCh <- i
			}()
			// ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
			// defer c()
			ctx := metadata.AppendToOutgoingContext(context.Background(), "authn-token", "a37bcba91de049aaa5ee1643472bf490")
			stream, err := ppsC.ListJobSet(ctx, &pps.ListJobSetRequest{
				Details: true,
			})
			if err != nil {
				atomic.AddInt64(&errs, 1)
				log.Printf("%d: start stream: %v", i, err)
				return
			}

			defer stream.CloseSend()
			atomic.AddInt64(&open, 1)
			log.Printf("%d: stream open", i)

			if _, err := stream.Recv(); err != nil {
				atomic.AddInt64(&errs, 1)
				log.Printf("%d: recv: %v", i, err)
				return
			}
			log.Printf("%d: got msg", i)
			atomic.AddInt64(&gotMsg, 1)
			time.Sleep(time.Hour)
		}(i)
	}
	go func() {
		for {
			time.Sleep(time.Second)
			log.Printf("%v open, %v got message, %v errored", atomic.LoadInt64(&open), atomic.LoadInt64(&gotMsg), atomic.LoadInt64(&errs))
		}
	}()
	for i := 1; i < *nstreams; i++ {
		streamID := <-doneCh
		log.Printf("stream %d done", streamID)
	}
}
