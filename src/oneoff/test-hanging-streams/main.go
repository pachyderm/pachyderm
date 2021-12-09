package main

import (
	"context"
	"log"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	cc, err := grpc.Dial("localhost:80", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	doneCh := make(chan int)
	ppsC := pps.NewAPIClient(cc)
	nstreams := 10
	for i := 0; i < nstreams; i++ {
		go func(i int) {
			defer func() {
				doneCh <- i
			}()
			ctx := metadata.AppendToOutgoingContext(context.Background(), "authn-token", "a37bcba91de049aaa5ee1643472bf490")
			stream, err := ppsC.ListJobSet(ctx, &pps.ListJobSetRequest{
				Details: true,
			})
			log.Printf("%d: stream open", i)
			if err != nil {
				log.Printf("%d: start stream: %v", i, err)
				return
			}
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("%d: recv: %v", i, err)
				return
			}
			log.Printf("%d: got msg: %v", i, msg.String())
			time.Sleep(time.Hour)
		}(i)
	}
	for i := 1; i < nstreams; i++ {
		streamID := <-doneCh
		log.Printf("stream %d done", streamID)
	}
}
