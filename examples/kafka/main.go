package main

import (
	"archive/tar"
	"context"
	"os"
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
		MinBytes: 10e3,
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

			// read a message
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				return err
			}
			// give it a unique name
			name := topic + time.Now().Format(time.RFC3339Nano)
			// write the header
			if err = tw.WriteHeader(&tar.Header{
				Name: name,
				Mode: 0600,
				Size: int64(len(m.Value)),
			}); err != nil {
				return err
			}
			// and the message
			if _, err = tw.Write(m.Value); err != nil {
				return err
			}
			return nil // this takes us back to the file loop

		}(); err != nil {
			panic(err)
		}
	}
}
