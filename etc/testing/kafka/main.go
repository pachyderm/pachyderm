package main

import (
	"archive/tar"
	"context"
	"os"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func main() {
	// Get the connection info from the ENV vars
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	topic := os.Getenv("TOPIC")

	// And create a new kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{host + ":" + port},
		Topic:    topic,
		MinBytes: 10e1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// Open the /pfs/out pipe with write only permissons (the pachyderm spout will be reading at the other end of this)
	// Note: it won't work if you try to open this with read, or read/write permissions
	out, err := os.OpenFile("/pfs/out", os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// this is the file loop
	for {
		if err := func() error {
			tw := tar.NewWriter(out)
			defer tw.Close()
			// this is the message loop
			// read a message
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				return err
			}
			// give it a unique name
			name := topic + time.Now().Format(time.RFC3339Nano)
			// write the header
			for err = tw.WriteHeader(&tar.Header{
				Name: name,
				Mode: 0600,
				Size: int64(len(m.Value)),
			}); err != nil; {
				if !strings.Contains(err.Error(), "broken pipe") {
					return err
				}
				// if there's a broken pipe, just give it some time to get ready for the next message
				time.Sleep(5 * time.Millisecond)
			}
			// and the message
			for _, err = tw.Write(m.Value); err != nil; {
				if !strings.Contains(err.Error(), "broken pipe") {
					return err
				}
				// if there's a broken pipe, just give it some time to get ready for the next message
				time.Sleep(5 * time.Millisecond)
			}
			return nil
		}(); err != nil {
			panic(err)
		}
	}
}
