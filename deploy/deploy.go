package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"text/template"
	"time"
)

var minPort, maxPort int = 49153, 65535 // Docker uses this range we're just replicating.

type service struct {
	Container, Name      string
	Shard, Nshards, Port int
}

func usage() {
	log.Print("Usage:")
	log.Print("$ deploy shard_count")
}

func printShardedService(name string) {
	sTemplate, err := template.New("sharded").ParseFiles("templates/sharded")
	if err != nil {
		log.Fatal(err)
	}
	aTemplate, err := template.New("announce_sharded").ParseFiles("templates/announce_sharded")
	if err != nil {
		log.Fatal(err)
	}

	for s := 0; s < *shards; s++ {
		config := new(service)
		config.Name = name
		config.Container = *container
		config.Shard = s
		config.Nshards = *shards
		config.Port = minPort + rand.Intn(maxPort-minPort)
		server, err := os.Create(fmt.Sprintf("%s-%d-%d.service", config.Name, config.Shard, config.Nshards))
		if err != nil {
			log.Fatal(err)
		}
		announce, err := os.Create(fmt.Sprintf("announce-%s-%d-%d.service", config.Name, config.Shard, config.Nshards))
		if err != nil {
			log.Fatal(err)
		}

		err = sTemplate.Execute(server, config)
		if err != nil {
			log.Fatal(err)
		}
		err = aTemplate.Execute(announce, config)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func printGlobalService(name string) {
	template, err := template.New("global").ParseFiles("templates/global")
	if err != nil {
		log.Fatal(err)
	}

	config := new(service)
	config.Name = name
	config.Container = *container
	config.Nshards = *shards

	server, err := os.Create(fmt.Sprintf("%s.service", config.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = template.Execute(server, config)
	if err != nil {
		log.Fatal(err)
	}
}

func printWebhookService(name string, port int) {
	template, err := template.New("webhook").ParseFiles("templates/webhook")
	if err != nil {
		log.Fatal(err)
	}

	config := new(service)
	config.Name = name
	config.Container = *container
	config.Port = port

	server, err := os.Create(fmt.Sprintf("%s.service", config.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = template.Execute(server, config)
	if err != nil {
		log.Fatal(err)
	}
}

func printRegistryService(name string, port int) {
	sTemplate, err := template.New("registry").ParseFiles("templates/registry")
	if err != nil {
		log.Fatal(err)
	}

	aTemplate, err := template.New("announce").ParseFiles("templates/announce")
	if err != nil {
		log.Fatal(err)
	}

	config := new(service)
	config.Name = name
	config.Port = port

	server, err := os.Create(fmt.Sprintf("%s.service", config.Name))
	if err != nil {
		log.Fatal(err)
	}

	announce, err := os.Create(fmt.Sprintf("announce-%s.service", config.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = sTemplate.Execute(server, config)
	if err != nil {
		log.Fatal(err)
	}

	err = aTemplate.Execute(announce, config)
	if err != nil {
		log.Fatal(err)
	}
}

func printGitDaemonService(name string) {
	template, err := template.New("gitdaemon").ParseFiles("templates/gitdaemon")
	if err != nil {
		log.Fatal(err)
	}

	config := new(service)
	config.Name = name

	server, err := os.Create(fmt.Sprintf("%s.service", config.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = template.Execute(server, config)
}

var shards *int
var container *string

func main() {
	log.SetFlags(log.Lshortfile)
	rand.Seed(time.Now().UTC().UnixNano())
	shards = flag.Int("shards", 3, "The number of shards in the deploy.")
	container = flag.String("container", "pachyderm/pfs", "The container to use for the deploy.")
	flag.Parse()

	printShardedService("shard")
	printGlobalService("router")
	printRegistryService("registry", 5000)
	printGitDaemonService("gitdaemon")
}
