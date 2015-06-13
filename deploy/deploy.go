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

var port = 49153

type service struct {
	Container, Name      string
	Shard, Nshards, Port int
	Disk                 string
}

var outPath string = "/host/home/core/pfs"

func printShardedService(name string) {
	sTemplate, err := template.New("sharded").ParseFiles("templates/sharded")
	if err != nil {
		log.Fatal(err)
	}

	for s := 0; s < *shards; s++ {
		for r := 0; r < *replicas; r++ {
			config := new(service)
			config.Name = name
			config.Container = *container
			config.Shard = s
			config.Nshards = *shards
			config.Port = port
			port++
			server, err := os.Create(fmt.Sprintf("%s/%s-%d-%d:%d.service", outPath, config.Name, config.Shard, config.Nshards, r))
			if err != nil {
				log.Fatal(err)
			}

			err = sTemplate.Execute(server, config)
			if err != nil {
				log.Fatal(err)
			}
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

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
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

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		log.Fatal(err)
	}

	announce, err := os.Create(fmt.Sprintf("%s/announce-%s.service", outPath, config.Name))
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

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = template.Execute(server, config)
}

func printStorageService(name string) {
	template, err := template.New("storage").ParseFiles("templates/storage")
	if err != nil {
		log.Fatal(err)
	}

	config := new(service)
	config.Name = name
	config.Disk = *disk

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = template.Execute(server, config)
}

var shards, replicas *int
var container *string
var disk *string

func main() {
	log.SetFlags(log.Lshortfile)
	rand.Seed(time.Now().UTC().UnixNano())
	shards = flag.Int("shards", 3, "The number of shards in the deploy.")
	replicas = flag.Int("replicas", 3, "The number of replicas of each shard.")
	container = flag.String("container", "pachyderm/pfs", "The container to use for the deploy.")
	disk = flag.String("disk", "/var/lib/pfs/data.img", "The disk to use for pfs' storage.")
	flag.Parse()

	printShardedService("shard")
	printGlobalService("router")
	printStorageService("storage")
}
