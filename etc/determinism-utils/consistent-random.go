package main

import (
	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

const (
	prngStateTable = "TEST_prngState"

	connectTimeoutSeconds = 5
	maxIdle               = 5
	maxOpen               = 100
)

func getCommand() *cobra.Command {
	var rethinkAddress string
	cmd = &cobra.Command{
		Use:   "consistent-random <id>",
		Short: "Generate a consistent sequence of pseudo-random numbers, for testing.",
		Long: `Generate a consistent sequence of pseudo-random numbers, for testing.

This binary connects to an existing RethinkDB instance and stores the state of
a PRNG in a row in RethinkDB keyed by <id>. Each time a test calls e.g.
'consistent-random <my-test-name>' it will emit the next number in the sequence.
This ensures that tests calling it will see the same sequence of random numbers
as long as the calls are serialized in the same way (as they would be when
testing restart behavior, for example).`,
		Run: pkgcobra.Bounds{Min: 1, Max: 1}, func(args []string) error {
			switch rethinkAddress {
			case "kube":
				rethinkAddress = "127.0.0.1:28015"
			case "localhost":
				rethinkAddress = "127.0.0.1:32081"
			}
			session, err := r.Connect(r.ConnectOpts{
				Address: rethinkAddress,
			})
			if err != nil {
				return fmt.Errorf("Could not connect to RethinkDB: \"%s\"", err.Error())
			}
			dbs, err := r.ListDB().Run(session)
			if err != nil {
				return fmt.Errorf("Could not get list of databases: \"%s\"", err.Error())
			}
			var list []string
			err = dbs.All(&list)
			if err != nil {
				return fmt.Errorf("dbs.All encountered an error: \"%s\"", err.Error())
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&rethinkAddress, "rethink-address", "kube",
		"host:port where RethinkDB can be reached. This can also be set to the special values "+
			"\"kube\" (which is designed to run inside a pachyderm pipeline, for testing, and "+
			"will use Pachyderm's RethinkDB instance) and \"localhost\" (which is designed to "+
			"run on your laptop, so you know what your tests will do). Default value is \"kube\".")
	return cmd
}

func connectToRethink()

func main() {
	err := getCommand().Execute()
	if err != nil {
		fmt.Errorf("Error: %s", err.String())
	}
}
