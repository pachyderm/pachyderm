package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func main() {
	// Get the connection info from the ENV vars
	topic := os.Getenv("TOPIC")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	// And create a new kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{host + ":" + port},
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// Open the /pfs/out pipe with write only permissons (the pachyderm spout will be reading at the other end of this)
	out, err := os.OpenFile("/pfs/out", os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprint("open", err))
	}
	defer out.Close()

	for {
		if err := func() error {
			tw := tar.NewWriter(out)
			defer tw.Close()

			var b bytes.Buffer

			tick := time.NewTicker(20 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				default:
					m, err := reader.ReadMessage(context.Background())
					if err != nil {
						panic(err)
					}
					b.Write(m.Value)
				case <-tick.C:
					if b.Len() == 0 {
						continue
					}
					name := topic + time.Now().Format(time.RFC3339Nano)
					if err = tw.WriteHeader(&tar.Header{
						Name: name,
						Mode: 0600,
						Size: int64(b.Len()),
					}); err != nil {
						return err
					}
					if _, err = b.WriteTo(tw); err != nil {
						return err
					}
					b.Reset()
					return nil
				}
			}
		}(); err != nil {
			panic(fmt.Sprint("outer loop", err))
		}
	}
}
