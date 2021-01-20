package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/streadway/amqp"

	client "github.com/pachyderm/pachyderm/src/client"
)

const defaultPrefetch = 500
const defaultFlushInterval = 10000 // ms
const defaultExtension = "ndjson"
const defaultBranch = "staging"
const defaultSwitchBranch = "master"
const defaultSwitchInterval = 60000 // ms

var logger log.Logger

func writeFiles(pc *client.APIClient, opts *options, buffer []amqp.Delivery, commits chan<- string) error {

	log.Print("writing buffer...")
	log.Printf("buffer_size %d", len(buffer))

	// hash messages to ensure there are no duplicated
	var maxMessage amqp.Delivery
	msgMap := make(map[string]amqp.Delivery)

	// groupby hash
	groupHash := md5.New()

	// Hash messages to ensure there are no duplicates
	for _, msg := range buffer {
		hash := md5.New()
		hash.Write(msg.Body)
		groupHash.Write(msg.Body)
		// Hash to ensure uniqueness of the messages
		name := hex.EncodeToString(hash.Sum(nil))

		// Keep track of the highest delivery tag
		if msg.DeliveryTag > maxMessage.DeliveryTag {
			maxMessage = msg
		}

		msgMap[name] = msg
	}

	// Give a warning if we found duplicates. Not a big deal
	if len(msgMap) != len(buffer) {
		log.Printf("duplicate messages in buffer: %d vs %d", len(msgMap), len(buffer))
	}

	// Start of writing
	log.Print("Writing messages...")

	// Begin the commit
	commit, err := pc.StartCommit(opts.repoName, opts.commitBranch)
	if err != nil {
		// StartCommit will fail if this crashed with an open commit. Finish open commits and
		commitList, err := pc.ListCommit(opts.repoName, opts.commitBranch, "", 0)
		for _, c := range commitList {
			com := c.GetCommit()
			err = pc.FinishCommit(opts.repoName, com.GetID())
			if err != nil {
				return fmt.Errorf("unable to start commit on %s@%s: %v", opts.repoName, opts.commitBranch, err)
			}
		}
		if err != nil {
			return fmt.Errorf("unable to start commit on %s@%s: %v", opts.repoName, opts.commitBranch, err)
		}
	}

	commitID := commit.GetID()

	// Name - to be used for groupby operations
	name := opts.topic + "-" + hex.EncodeToString(groupHash.Sum(nil)) + "." + opts.ext

	reader, writer := io.Pipe()

	errors := make(chan error, 1)
	// Write messages into a pip so they can be read by pach client
	go func() {
		for _, msg := range msgMap {
			// Write each message followed by newline - you may need to do initial parsing
			// for safety here if your messages have newline characters or change this up entirely.
			writer.Write(msg.Body)
			writer.Write([]byte("\n"))
		}
		// Close writer. Log errors.
		if err := writer.Close(); err != nil {
			errors <- err
		}
	}()

	// Write to PFS.
	if opts.overwrite {
		_, err := pc.PutFileOverwrite(opts.repoName, commitID, name, reader, 0)
		if err != nil {
			return err
		}
	} else {
		_, err := pc.PutFile(opts.repoName, commitID, name, reader)
		if err != nil {
			return err
		}
	}

	// Was there an error? If so, send back to main since we aren't handling it.
	select {
	case err := <-errors:
		return err
	default:
	}

	// Finish the commit
	err = pc.FinishCommit(opts.repoName, commitID)
	if err != nil {
		return fmt.Errorf("unable to finish commit on %s@%s:%s: %v", opts.repoName, opts.commitBranch, commit.GetID(), err)
	}

	// Write latest finished commit to channel
	commits <- commitID

	log.Print("wrote messages.")

	// If test mode is on, this "nacks" and requeues messages so you don't have to keep filling up your test queue.
	if opts.test {
		log.Print("Nack")
		maxMessage.Nack(true, true)
	} else {
		// Bulk acknowledge all messages writen since last flush.
		log.Print("Ack")
		maxMessage.Ack(true)

	}

	// yay
	return nil
}

func bufferReceive(pc *client.APIClient, msgC <-chan amqp.Delivery, opts *options, commits chan<- string, errors chan<- error) {
	// Flush the read messages once every
	flushC := time.NewTicker(time.Duration(opts.flushInterval) * time.Millisecond)
	buffer := make([]amqp.Delivery, 0, cap(msgC))

	for {
		select {
		case msg := <-msgC: // New message to read
			buffer = append(buffer, msg)
			// Check for bull buffer
			if len(buffer) == cap(buffer) {

				err := writeFiles(pc, opts, buffer, commits)
				if err != nil {
					errors <- err
				}
				buffer = buffer[:0]
				//flushC = time.NewTicker(time.Duration(flushInterval) * time.Millisecond)
				flushC.Reset(time.Duration(opts.flushInterval) * time.Millisecond)
			}
		case <-flushC.C: // On a timer
			if len(buffer) > 0 {
				// Flush the buffer
				err := writeFiles(pc, opts, buffer, commits)
				if err != nil {
					errors <- err
				}
				buffer = buffer[:0]
			}

		}
	}
}

func switchBranch(pc *client.APIClient, opts *options, commits <-chan string, errors chan<- error) {

	log.Print("started branch switching routine.")
	switchC := time.NewTicker(time.Duration(opts.switchInterval) * time.Millisecond)

	latestClosedCommit := ""

	// This reads closed commits until the branch switch interval is reached and then sets master
	for {
		select {
		case latestClosedCommit = <-commits:
			log.Printf("latest commit %s", latestClosedCommit)
		case <-switchC.C:
			if latestClosedCommit != "" {
				log.Printf("switching branch %s HEAD to %s", opts.switchBranch, latestClosedCommit)
				err := pc.CreateBranch(opts.repoName, opts.switchBranch, latestClosedCommit, nil)
				if err != nil {
					errors <- err
					return
				}
			}
		}
	}
}

func consume(pc *client.APIClient, msgs <-chan amqp.Delivery, opts *options, commits chan<- string, errors chan<- error) {

	// Messages to write to file.
	msgC := make(chan amqp.Delivery, opts.prefetch)

	// Simple goroutine to read incoming messages from a channel
	go bufferReceive(pc, msgC, opts, commits, errors)

	// This may look funny, but the msgs channel is unbuffered. I don't like debugging channel deadlocks, do you?
	for msg := range msgs {
		// Write message to the buffer
		msgC <- msg

	}

}

type options struct {
	prefetch       int
	flushInterval  int
	switchInterval int
	topic          string
	err            error
	test           bool
	overwrite      bool
	repoName       string
	commitBranch   string
	switchBranch   string
	ext            string
}

func parseOptions(opts *options) {
	var err error

	// Prefetch
	prefetchStr, ok := os.LookupEnv("PREFETCH")
	if ok {
		opts.prefetch, err = strconv.Atoi(prefetchStr)
		if err != nil {
			log.Fatal("unable to parse PREFETCH")
		}
	}

	// Flush Interval
	flushIntStr, ok := os.LookupEnv("FLUSH_INTERVAL_MS")
	if ok {
		opts.flushInterval, err = strconv.Atoi(flushIntStr)
		if err != nil {
			log.Fatal("unable to parse FLUSH_INTERVAL_MS")
		}
	}

	// Branch Switch Interval
	switchIntStr, ok := os.LookupEnv("SWITCH_INTERVAL_MS")
	if ok {
		opts.switchInterval, err = strconv.Atoi(switchIntStr)
		if err != nil {
			log.Fatal("unable to parse SWITCH_INTERVAL_MS")
		}
	}

	// File Extension
	extStr, ok := os.LookupEnv("EXTENSION")
	if ok {
		opts.ext = extStr
	}

	testStr, ok := os.LookupEnv("TEST")
	if ok {
		opts.test, err = strconv.ParseBool(testStr)
		if err != nil {
			log.Fatal("unable to parse TEST")
		}
	}

	switchBranchName, ok := os.LookupEnv("SWITCH_BRANCH")
	if ok {
		opts.switchBranch = switchBranchName
	}

	commitBranchName, ok := os.LookupEnv("COMMIT_BRANCH")
	if ok {
		opts.commitBranch = commitBranchName
	}

	// Topic/Queue for this to read from
	opts.topic = os.Getenv("TOPIC")

	opts.repoName = os.Getenv("PPS_PIPELINE_NAME")
}

func dialRabbit() (*amqp.Connection, error) {
	rabbitURL := os.Getenv("RABBITMQ_HOST")
	if rabbitURL == "" {
		return nil, fmt.Errorf("RABBITMQ_HOST environment variable not set")
	}

	rabbitUser := os.Getenv("RABBITMQ_USER")
	if rabbitUser == "" {
		return nil, fmt.Errorf("RABBITMQ_USER environment variable not set")
	}
	rabbitPassword := os.Getenv("RABBITMQ_PASSWORD")
	if rabbitPassword == "" {
		return nil, fmt.Errorf("RABBITMQ_PASSWORD environment variable not set")
	}

	return amqp.Dial("amqp://" + rabbitUser + ":" + rabbitPassword + "@" + rabbitURL)
}

func main() {
	// Buffered channel for errors. 50 should do it.
	errors := make(chan error, 50)
	// Buffered channel of latest commit
	commits := make(chan string, 1)

	// Create a Pachd client
	pc, err := client.NewInWorker()
	if err != nil {
		log.Fatalf("Unable to create pachd client: %v", err)
		return
	}
	// Close if this exits, log errors if closing fails
	defer func() {
		err := pc.Close()
		if err != nil {
			errors <- err
		}
	}()

	// Option defaults
	opts := options{
		prefetch:      defaultPrefetch,
		flushInterval: defaultFlushInterval,
		ext:           defaultExtension,
		test:          false,
	}

	parseOptions(&opts)

	flag.StringVar(&opts.topic, "topic", "", "RabbitMQ Topic")
	flag.BoolVar(&opts.overwrite, "overwrite", true, "overwrite existing files")

	flag.Parse()

	if opts.topic == "" {
		log.Fatal("topic to read from not specified")
	}

	// Dial rabbit to get a connection
	conn, err := dialRabbit()
	if err != nil {
		log.Fatalf("unable to dial RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a channel.
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("unable to open a channel: %v", err)
	}
	defer channel.Close()

	// Set prefetch (max messages unacknowledged)
	channel.Qos(opts.prefetch, 0, false)

	// Declare queue (idempotent. Safe even if you know queue exists)
	q, err := channel.QueueDeclare(
		opts.topic,
		true,  // durable
		false, // auto delete
		false, // non-exclusive
		false, // wait for queue
		nil,   // no args
	)
	if err != nil {
		log.Fatalf("unable to create or access queue with name '%s': %v", opts.topic, err)
	}

	// Finally, consume from the queue. This is a channel of amqp.Delivery
	msgs, err := channel.Consume(
		q.Name,
		"",    // use primary exchange
		false, // don't autoack
		false, // not exclusive (can be multiple consumers)
		false, // *does nothing?*
		false, // wait for server to confirm
		nil,   // no args
	)
	if err != nil {
		log.Fatalf("unable to consume on channel for queue '%s': %v", q.Name, err)
	}

	// Switch branch on a timer
	go switchBranch(pc, &opts, commits, errors)
	// Consume from RabbitMQ
	go consume(pc, msgs, &opts, commits, errors)

	// Log any errors.
	log.Fatal(<-errors)

}
