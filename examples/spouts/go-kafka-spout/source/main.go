package main

import (
	"archive/tar"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

const defaultHost = "kafka.kafka"
const defaultPort = "9092"
const defaultTopic = "test"
const defaultGroupID = "test"
const defaultTimeout = 5
const defaultNamedPipe = "/pfs/out"

func process_messages(pipe string, reader *kafka.Reader, timeout int, log bool) error {
	// read a message
	if log {
		fmt.Printf("reading kafka queue.\n")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer func() {
		if log {
			fmt.Printf("cleaning up context.\n")
		}
		cancel()
	}()
	m, err := reader.ReadMessage(ctx)
	if err != nil {
		return err
	}
	if log {
		fmt.Printf("opening named pipe %v.\n", pipe)
	}
	// Open the /pfs/out pipe with write only permissons (the pachyderm spout will be reading at the other end of this)
	// Note: it won't work if you try to open this with read, or read/write permissions
	out, err := os.OpenFile(pipe, os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		if log {
			fmt.Printf("closing named pipe %v.\n", pipe)
		}
		out.Close()
	}()

	if log {
		fmt.Printf("opening tarstream\n")
	}
	tw := tar.NewWriter(out)
	defer func() {
		if log {
			fmt.Printf("closing tarstream.\n")
		}
		tw.Close()
	}()

	if log {
		fmt.Printf("processing header for topic %v @ offset %v\n", m.Topic, m.Offset)
	}
	// give it a unique name
	name := fmt.Sprintf("%v-%v", m.Topic, m.Offset)
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
		if log {
			fmt.Printf("broken pipe\n")
		}
		time.Sleep(time.Duration(timeout) * time.Millisecond)
	}
	if log {
		fmt.Printf("processing data for topic  %v @ offset %v\n", m.Topic, m.Offset)
	}
	// and the message
	for _, err = tw.Write(m.Value); err != nil; {
		if !strings.Contains(err.Error(), "broken pipe") {
			return err
		}
		// if there's a broken pipe, just give it some time to get ready for the next message
		if log {
			fmt.Printf("broken pipe\n")
		}
		time.Sleep(time.Duration(timeout) * time.Millisecond)
	}
	return nil
}

func main() {

	// Set the default values of the configurable variables
	var (
		host       = defaultHost
		port       = defaultPort
		topic      = defaultTopic
		group_id   = defaultGroupID
		timeout    = defaultTimeout
		named_pipe = defaultNamedPipe
		log        = false
		ok         = false
		tmp        = ""
	)

	// override with environment variables if they're set
	host, ok = os.LookupEnv("KAFKA_HOST")
	if !ok {
		host = defaultHost
	}
	port, ok = os.LookupEnv("KAFKA_PORT")
	if !ok {
		port = defaultPort
	}
	topic, ok = os.LookupEnv("KAFKA_TOPIC")
	if !ok {
		topic = defaultTopic
	}
	group_id, ok = os.LookupEnv("KAFKA_GROUP_ID")
	if !ok {
		group_id = defaultGroupID
	}
	tmp, ok = os.LookupEnv("KAFKA_TIMEOUT")
	if ok {
		if myTimeout, err := strconv.Atoi(tmp); err != nil {
			fmt.Sprintf("error parsing KAFKA_TIMEOUT env value %v, using default %v unless flagged\n", tmp, defaultTimeout)
			timeout = defaultTimeout
		} else {
			timeout = myTimeout
		}
	} else {
		timeout = defaultTimeout

	}
	named_pipe, ok = os.LookupEnv("NAMED_PIPE")
	if !ok {
		named_pipe = defaultNamedPipe
	}
	tmp, ok = os.LookupEnv("VERBOSE_LOGGING")
	if !ok || strings.EqualFold("false", tmp) || tmp == "" {
		log = false
	} else {
		log = true
	}

	// override with flags if they're set
	flag.StringVar(&host, "kafka_host", host, "the hostname of the Kafka broker")
	flag.StringVar(&port, "kafka_port", port, "the port of the Kafka broker")
	flag.StringVar(&topic, "kafka_topic", topic, "the Kafka topic for messages")
	flag.StringVar(&group_id, "kafka_group_id", group_id, "the Kafka group for maintaining offset state")
	flag.IntVar(&timeout, "kafka_timeout", timeout, "the timeout in seconds for reading messages from the Kafka queue")
	flag.StringVar(&named_pipe, "named_pipe", named_pipe, "the named pipe for the spout")
	flag.BoolVar(&log, "v", log, "verbose logging")

	flag.Parse()

	// And create a new kafka reader
	if log {
		fmt.Printf("creating new kafka reader for %v:%v with topic '%v' and group '%v'\n", host, port, topic, group_id)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{host + ":" + port},
		Topic:    topic,
		GroupID:  group_id,
		MinBytes: 10e1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	for {
		err := process_messages(named_pipe, reader, timeout, log)
		if err != nil {
			if log {
				if !strings.Contains(err.Error(), "context deadline exceeded") {
					fmt.Printf("error processing kafka: %v\n", err)
				} else {
					fmt.Printf("timeout\n")
				}
			}
		}
		// fix for https://github.com/pachyderm/pachyderm/issues/4327
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
}
