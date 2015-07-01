package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/pachyderm/pachyderm/src/log"
)

var port = 49153

type service struct {
	Container, Name      string
	Shard, Nshards, Port int
	Disk                 string
}

var outPath string = "/host/home/core/pfs"

func printShardedService(name string) error {
	sTemplate, err := template.New("sharded").Parse(shardedTemplateString)
	if err != nil {
		return err
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
				return err
			}

			err = sTemplate.Execute(server, config)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func printGlobalService(name string) error {
	template, err := template.New("global").Parse(globalTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name
	config.Container = *container
	config.Nshards = *shards

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	return template.Execute(server, config)
}

func printRegistryService(name string, port int) error {
	sTemplate, err := template.New("registry").Parse(registryTemplateString)
	if err != nil {
		return err
	}

	aTemplate, err := template.New("announce").Parse(announceTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name
	config.Port = port

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	announce, err := os.Create(fmt.Sprintf("%s/announce-%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	err = sTemplate.Execute(server, config)
	if err != nil {
		return err
	}

	return aTemplate.Execute(announce, config)
}

func printGitDaemonService(name string) error {
	template, err := template.New("gitdaemon").Parse(gitDaemonTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	return template.Execute(server, config)
}

func printStorageService(name string) error {
	template, err := template.New("storage").Parse(storageTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name
	config.Disk = *disk

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	return template.Execute(server, config)
}

var shards, replicas *int
var container *string
var disk *string

func main() {
	if err := do(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	rand.Seed(time.Now().UTC().UnixNano())
	shards = flag.Int("shards", 3, "The number of shards in the deploy.")
	replicas = flag.Int("replicas", 3, "The number of replicas of each shard.")
	container = flag.String("container", "pachyderm/pfs", "The container to use for the deploy.")
	disk = flag.String("disk", "/var/lib/pfs/data.img", "The disk to use for pfs' storage.")
	flag.Parse()

	if err := printShardedService("shard"); err != nil {
		return err
	}
	if err := printGlobalService("router"); err != nil {
		return err
	}
	if err := printStorageService("storage"); err != nil {
		return err
	}
	return nil
}
