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
	if len(args) != 4 {
		log.Fatalf("Expected two arguments, input/output yaml file & pfs dir, got %d", len(args)-1)
	}

	// Pre-process the provided manifest.
	tfJob, err := preprocessManifest(args)
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
func preprocessManifest(args []string) (*TFJob, error) {

	// Decode the YAML file.
	var tfJob TFJob

	f, err := ioutil.ReadFile(args[1])
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(f, &tfJob); err != nil {
		return nil, err
	}

	// Output the job name to stdout.
	fmt.Printf("%s", tfJob.Metadata.Name)

	// TODO further mod the manifest (sidecars, etc.)

	// Add input repos as environmental vars.
	if err := addInputRepos(&tfJob, args[3]); err != nil {
		return nil, err
	}

	return &tfJob, nil
}

// addInputRepos adds the PFS input repos to the TFJob manifest
// as environmental variables.
func addInputRepos(tfJob *TFJob, pfsDir string) error {

	// Collect input repo names.
	var repos []string

	// Walk over the repos in PFS.
	if err := filepath.Walk(pfsDir, func(path string, info os.FileInfo, err error) error {

		// Process any directory under /pfs.
		if info.IsDir() {

			// Skip if it's the output repo.
			if info.Name() == "out" || info.Name() == "pfs" {
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
	for repID, replica := range tfJob.Spec.ReplicaSpecs {
		for conID, container := range replica.Template.Spec.Containers {
			tfJob.Spec.ReplicaSpecs[repID].Template.Spec.Containers[conID].Envs = append(container.Envs, reposJoinedEnv)
		}
	}

	return nil
}
