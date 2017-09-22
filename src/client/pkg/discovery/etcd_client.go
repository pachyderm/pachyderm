package discovery

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

type etcdClient struct {
	client *etcd.Client
}

// customCheckRetry is a fork of etcd's DefaultCheckRetry, except that it issues
// more retries before giving up. Because Pachyderm often starts before etcd is
// ready, retrying Pachd's connection to etcd in a tight loop (<1s) is often
// much faster than waiting for kubernetes to restart the pachd pod.
func customCheckRetry(cluster *etcd.Cluster, numReqs int, lastResp http.Response,
	err error) error {
	// Retry for 5 minutes, unless the cluster is super huge
	maxRetries := 2 * len(cluster.Machines)
	if 600 > maxRetries {
		maxRetries = 600
	}
	if numReqs > maxRetries {
		errStr := fmt.Sprintf("failed to propose on members %v [last error: %v]", cluster.Machines, err)
		return &etcd.EtcdError{
			ErrorCode: etcd.ErrCodeEtcdNotReachable,
			Message:   "All the given peers are not reachable",
			Cause:     errStr,
			Index:     0,
		}
	}

	if lastResp.StatusCode == 0 {
		// always retry if it failed to get a response
		return nil
	}
	if lastResp.StatusCode != http.StatusInternalServerError {
		// The status code  indicates that etcd is no longer in leader election.
		// Something is wrong
		body := []byte("nil")
		if lastResp.Body != nil {
			if b, err := ioutil.ReadAll(lastResp.Body); err == nil {
				body = b
			}
		}
		errStr := fmt.Sprintf("unhandled http status [%s] with body [%s]", http.StatusText(lastResp.StatusCode), body)
		return &etcd.EtcdError{
			ErrorCode: etcd.ErrCodeUnhandledHTTPStatus,
			Message:   "Unhandled HTTP Status",
			Cause:     errStr,
			Index:     0,
		}
	}

	// sleep some time and expect leader election finish
	time.Sleep(time.Millisecond * 500)
	fmt.Println("Warning: bad response status code ", lastResp.StatusCode)
	return nil
}

func newEtcdClient(addresses ...string) *etcdClient {
	client := etcd.NewClient(addresses)
	client.CheckRetry = customCheckRetry
	return &etcdClient{client}
}

func (c *etcdClient) Close() error {
	c.client.Close()
	return nil
}

func (c *etcdClient) Get(key string) (string, error) {
	response, err := c.client.Get(key, false, false)
	if err != nil {
		return "", err
	}
	return response.Node.Value, nil
}

func (c *etcdClient) GetAll(key string) (map[string]string, error) {
	response, err := c.client.Get(key, false, true)
	result := make(map[string]string, 0)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			return result, nil
		}
		return nil, err
	}
	nodeToMap(response.Node, result)
	return result, nil
}

func (c *etcdClient) Watch(key string, cancel chan bool, callBack func(string) error) error {
	// This retry is needed for when the etcd cluster gets overloaded.
	for {
		if err := c.watchWithoutRetry(key, cancel, callBack); err != nil {
			etcdErr, ok := err.(*etcd.EtcdError)
			if ok && etcdErr.ErrorCode == 401 {
				continue
			}
			if ok && etcdErr.ErrorCode == 501 {
				continue
			}
			return err
		}
	}
}

func (c *etcdClient) WatchAll(key string, cancel chan bool, callBack func(map[string]string) error) error {
	for {
		if err := c.watchAllWithoutRetry(key, cancel, callBack); err != nil {
			etcdErr, ok := err.(*etcd.EtcdError)
			if ok && etcdErr.ErrorCode == 401 {
				continue
			}
			if ok && etcdErr.ErrorCode == 501 {
				continue
			}
			return err
		}
	}
}

func (c *etcdClient) Set(key string, value string, ttl uint64) error {
	_, err := c.client.Set(key, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) Create(key string, value string, ttl uint64) error {
	_, err := c.client.Create(key, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) CreateInDir(dir string, value string, ttl uint64) error {
	_, err := c.client.CreateInOrder(dir, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) Delete(key string) error {
	_, err := c.client.Delete(key, false)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) CheckAndDelete(key string, oldValue string) error {
	_, err := c.client.CompareAndDelete(key, oldValue, 0)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) CheckAndSet(key string, value string, ttl uint64, oldValue string) error {
	var err error
	if oldValue == "" {
		_, err = c.client.Create(key, value, ttl)
	} else {
		_, err = c.client.CompareAndSwap(key, value, ttl, oldValue, 0)
	}
	if err != nil {
		return err
	}
	return nil
}

// nodeToMap translates the contents of a node into a map
// nodeToMap can be called on the same map with successive results from watch
// to accumulate a value
// nodeToMap returns true if out was modified
func nodeToMap(node *etcd.Node, out map[string]string) bool {
	key := strings.TrimPrefix(node.Key, "/")
	if !node.Dir {
		if node.Value == "" {
			if _, ok := out[key]; ok {
				delete(out, key)
				return true
			}
			return false
		}
		if value, ok := out[key]; !ok || value != node.Value {
			out[key] = node.Value
			return true
		}
		return false
	}
	changed := false
	for _, node := range node.Nodes {
		changed = nodeToMap(node, out) || changed
	}
	return changed
}

func (c *etcdClient) watchWithoutRetry(key string, cancel chan bool, callBack func(string) error) error {
	var waitIndex uint64 = 1
	// First get the starting value of the key
	response, err := c.client.Get(key, false, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			err = callBack("")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		err = callBack(response.Node.Value)
		if err != nil {
			return err
		}
		waitIndex = response.Node.ModifiedIndex + 1
	}
	for {
		response, err := c.client.Watch(key, waitIndex, false, nil, cancel)
		if err != nil {
			if err == etcd.ErrWatchStoppedByUser {
				return ErrCancelled
			}
			return err
		}
		err = callBack(response.Node.Value)
		if err != nil {
			return err
		}
		waitIndex = response.Node.ModifiedIndex + 1
	}
}

func (c *etcdClient) watchAllWithoutRetry(key string, cancel chan bool, callBack func(map[string]string) error) error {
	var waitIndex uint64 = 1
	value := make(map[string]string)
	// First get the starting value of the key
	response, err := c.client.Get(key, false, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			err = callBack(nil)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		waitIndex = response.EtcdIndex + 1
		if nodeToMap(response.Node, value) {
			err = callBack(value)
			if err != nil {
				return err
			}
		}
	}
	for {
		response, err := c.client.Watch(key, waitIndex, true, nil, cancel)
		if err != nil {
			if err == etcd.ErrWatchStoppedByUser {
				return ErrCancelled
			}
			return err
		}
		waitIndex = response.EtcdIndex + 1
		if nodeToMap(response.Node, value) {
			err = callBack(value)
			if err != nil {
				return err
			}
		}
	}
}
