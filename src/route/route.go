package route

import (
	"bytes"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/coreos/go-etcd/etcd"
	"github.com/pachyderm/pachyderm/src/etcache"
)

var ErrNoHosts = errors.New("pfs: no hosts found")

func HashResource(resource string) uint64 {
	return uint64(adler32.Checksum([]byte(resource)))
}

// Parse a string descriving a shard, the string looks like: "0-4"
func ParseShard(shardDesc string) (uint64, uint64, error) {
	s_m := strings.Split(shardDesc, "-")
	shard, err := strconv.ParseUint(s_m[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	modulos, err := strconv.ParseUint(s_m[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return shard, modulos, nil
}

// Match returns true of a resource hashes to the given shard.
func Match(resource, shardDesc string) (bool, error) {
	shard, modulos, err := ParseShard(shardDesc)
	if err != nil {
		return false, err
	}
	return (HashResource(resource) % modulos) == shard, nil
}

func Route(cache etcache.Cache, r *http.Request, etcdKey string, modulos uint64) (io.ReadCloser, error) {
	bucket := hashRequest(r) % modulos
	shard := fmt.Sprint(bucket, "-", fmt.Sprint(modulos))

	_master, err := cache.Get(path.Join(etcdKey, shard), false, false)
	if err != nil {
		return nil, err
	}
	master := _master.Node.Value

	httpClient := &http.Client{}
	// `Do` will complain if r.RequestURI is set so we unset it
	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Host = strings.TrimPrefix(master, "http://")
	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed request (%s) to %s.", resp.Status, r.URL.String())
	}
	return resp.Body, nil
}

func RouteHttp(cache etcache.Cache, w http.ResponseWriter, r *http.Request, etcdKey string, modulos uint64) {
	reader, err := Route(cache, r, etcdKey, modulos)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// Multicast enables the Ogre Magi to rapidly cast his spells, giving them
// greater potency.
// Multicast sends a request to every host it finds under a key and returns a
// ReadCloser for each one.
func Multicast(cache etcache.Cache, r *http.Request, etcdKey string) ([]*http.Response, error) {
	_endpoints, err := cache.Get(etcdKey, false, true)
	if err != nil {
		return nil, err
	}
	endpoints := _endpoints.Node.Nodes
	if len(endpoints) == 0 {
		return nil, ErrNoHosts
	}

	// If the request has a body we need to store it in memory because it needs
	// to be sent to multiple endpoints and Reader (the type of r.Body) is
	// single use.
	var body []byte
	if r.ContentLength != 0 {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}

	var resps []*http.Response
	errors := make(chan error, len(endpoints))
	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	for _, node := range endpoints {
		go func(node *etcd.Node) {
			defer wg.Done()
			httpClient := &http.Client{}
			// First make a request, taking some values from the previous request.
			url := node.Value + r.URL.Path + "?" + r.URL.RawQuery
			req, err := http.NewRequest(r.Method, url,
				ioutil.NopCloser(bytes.NewReader(body)))
			if err != nil {
				errors <- err
				return
			}
			// Send the request
			resp, err := httpClient.Do(req)
			if err != nil {
				errors <- err
				return
			}
			if resp.StatusCode != 200 {
				errors <- fmt.Errorf("Failed request (%s) to %s.", resp.Status, r.URL.String())
				return
			}
			// Append the request to the response slice.
			lock.Lock()
			resps = append(resps, resp)
			lock.Unlock()
		}(node)
	}
	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return resps, nil
}

type Return int

const (
	// ReturnFirst returns only the first response
	ReturnOne Return = iota
	// ReturnAll returns all the responses
	ReturnAll Return = iota
)

// MulticastHttp sends r to every host it finds under etcdKey, then prints the
// response to w based on
func MulticastHttp(cache etcache.Cache, w http.ResponseWriter, r *http.Request, etcdKey string, ret Return) {
	// resps is guaranteed to be nonempty
	resps, err := Multicast(cache, r, etcdKey)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer func() {
		for _, r := range resps {
			r.Body.Close()
		}
	}()
	switch ret {
	case ReturnOne:
		_, err = io.Copy(w, resps[0].Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		return
	case ReturnAll:
		// We use the existence of "Boundary" to figure out how to concatenate
		// the responses
		if resps[0].Header.Get("Boundary") == "" {
			// plain text
			for _, resp := range resps {
				_, err = io.Copy(w, resp.Body)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				resp.Body.Close()
			}
		} else {
			// multipart
			writer := multipart.NewWriter(w)
			defer writer.Close()
			w.Header().Add("Boundary", writer.Boundary())
			for _, resp := range resps {
				reader := multipart.NewReader(resp.Body, resp.Header.Get("Boundary"))
				for p, err := reader.NextPart(); err == nil; p, err = reader.NextPart() {
					f, err := writer.CreateFormFile(p.FormName(), p.FileName())
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
					_, err = io.Copy(f, p)
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
				}
				resp.Body.Close()
			}
		}
	}
}

func hashRequest(r *http.Request) uint64 {
	return HashResource(r.URL.Path)
}
