package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/pachyderm/pachyderm/src/client"
)

// Env contains the basic setup for running end-to-end pachyderm tests entirely
// locally within the test process. It provides a temporary directory for
// storing data, an embedded etcd server with a connected client, as well as a
// local mock pachd instance which allows a test to hook into any pachd calls.
type Env struct {
	Directory  string
	Etcd       *embed.Etcd
	EtcdClient *etcd.Client
	MockPachd  *MockPachd
	PachClient *client.APIClient
}

// WithEnv sets up an Env structure, passes it to the provided callback, then
// cleans up everything in the environment, regardless of if an assertion fails.
func WithEnv(cb func(*Env) error) (err error) {
	env := &Env{}

	dirBase := path.Join(os.TempDir(), "pachyderm_test")

	err = os.MkdirAll(dirBase, 0700)
	if err != nil {
		return err
	}

	env.Directory, err = ioutil.TempDir(dirBase, "")
	if err != nil {
		return err
	}

	// Cleanup any state when we return
	defer func() {
		fmt.Printf("Cleaning up testutil.Env, err: %v\n", err)
		// We return the first error that occurs during teardown, but still try to
		// close everything
		saveErr := func(e error) error {
			if e != nil && err == nil {
				err = e
			}
			return e
		}

		if env.PachClient != nil {
			e := saveErr(env.PachClient.Close())
			fmt.Printf("Closed PachClient, err: %v\n", e)
		}

		if env.MockPachd != nil {
			e := saveErr(env.MockPachd.Close())
			fmt.Printf("Closed MockPachd, err: %v\n", e)
		}

		if env.EtcdClient != nil {
			e := saveErr(env.EtcdClient.Close())
			fmt.Printf("Closed EtcdClient, err: %v\n", e)
		}

		if env.Etcd != nil {
			env.Etcd.Close()
		}

		e := saveErr(os.RemoveAll(env.Directory))
		fmt.Printf("Removed temp directory, err: %v\n", e)
	}()

	etcdConfig := embed.NewConfig()

	// Create test dirs for etcd data
	etcdConfig.Dir, err = ioutil.TempDir(env.Directory, "etcd_data")
	if err != nil {
		return err
	}
	etcdConfig.WalDir, err = ioutil.TempDir(env.Directory, "etcd_wal")
	if err != nil {
		return err
	}

	// Speed up initial election, hopefully this has no other impact since there
	// is only one etcd instance
	etcdConfig.InitialElectionTickAdvance = true
	etcdConfig.TickMs = 2
	etcdConfig.ElectionMs = 10

	env.Etcd, err = embed.StartEtcd(etcdConfig)
	if err != nil {
		return err
	}

	clientUrls := []string{}
	for _, url := range etcdConfig.LCUrls {
		clientUrls = append(clientUrls, url.String())
	}

	env.EtcdClient, err = etcd.New(etcd.Config{
		Endpoints:   clientUrls,
		DialOptions: client.DefaultDialOptions(),
	})
	if err != nil {
		return err
	}

	env.MockPachd = NewMockPachd()

	env.PachClient, err = client.NewFromAddress(env.MockPachd.Addr.String())
	if err != nil {
		return err
	}

	return cb(env)
}
