package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const reposEnvVar = "PACH_REPOS"

// Container include info about a container in a pod.
type Container struct {
	Image string `yaml:"image"`
	Name  string `yaml:"name"`
	Envs  []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"env"`
}

// ReplicaSpec includes info about a replica set.
type ReplicaSpec struct {
	Replicas      int    `yaml:"replicas"`
	TFPort        string `yaml:"tfPort,omitempty"`
	TFReplicaType string `yaml:"tfReplicaType"`
	Template      struct {
		Spec struct {
			Containers    []Container `yaml:"containers"`
			RestartPolicy string      `yaml:"restartPolicy"`
		} `yaml:"spec"`
	} `yaml:"template"`
}

// TFJob contains all of the information utilized to create a TFJob.
type TFJob struct {
	APIVersion string                `yaml:"apiVersion"`
	Kind       string                `yaml:"kind"`
	Metadata   struct{ Name string } `yaml:"metadata"`
	Spec       struct {
		ReplicaSpecs []ReplicaSpec `yaml:"replicaSpecs"`
	} `yaml:"spec"`
}

func main() {

	// Parse the command line arguments.
	args := os.Args
	if len(args) != 3 {
		log.Fatalf("Expected two arguments, input and output yaml file, got %d", len(args)-1)
	}

	// Pre-process the provided manifest.
	tfJob, err := preprocessManifest(args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Output the pre-processed manifest.
	tfJobOut, err := yaml.Marshal(tfJob)
	if err := ioutil.WriteFile(args[2], tfJobOut, 0644); err != nil {
		log.Fatal(err)
	}
}

// preprocessManifest pre-processing an input YAML manifest.
func preprocessManifest(filename string) (*TFJob, error) {

	// Decode the YAML file.
	var tfJob TFJob

	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(f, &tfJob); err != nil {
		return nil, err
	}

	// TODO further mod the manifest (sidecars, etc.)

	// Add input repos as environmental vars.
	if err := addInputRepos(&tfJob); err != nil {
		return nil, err
	}

	return &tfJob, nil
}

// addInputRepos adds the PFS input repos to the TFJob manifest
// as environmental variables.
func addInputRepos(tfJob *TFJob) error {

	// Collect input repo names.
	var repos []string

	// Walk over the repos in PFS.
	if err := filepath.Walk("/pfs", func(path string, info os.FileInfo, err error) error {
		fmt.Println(info.Name())

		// Process any directory under /pfs.
		if info.IsDir() {

			// Skip if it's the output repo.
			if info.Name() == "out" {
				return nil
			}

			// Otherwise, add the repo name to the
			// list of repo names.
			repos = append(repos, info.Name())
		}

		return nil
	}); err != nil {
		return err
	}

	// Join the repo names and add env vars to the
	// TFJob containers.
	reposJoined := strings.Join(repos, ",")
	reposJoinedEnv := struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	}{
		Name:  reposEnvVar,
		Value: reposJoined,
	}
	for _, replica := range tfJob.Spec.ReplicaSpecs {
		for _, container := range replica.Template.Spec.Containers {
			container.Envs = append(container.Envs, reposJoinedEnv)
		}
	}

	return nil
}
