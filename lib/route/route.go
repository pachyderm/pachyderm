package route

import (
	"bytes"
	"fmt"
	"hash/adler32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/pachyderm/pfs/lib/etcache"
)

func HashResource(resource string) uint64 {
	return uint64(adler32.Checksum([]byte(resource)))
}

func hashRequest(r *http.Request) uint64 {
	return HashResource(r.URL.Path)
}

func Route(r *http.Request, etcdKey string, modulos uint64) (io.ReadCloser, error) {
	bucket := hashRequest(r) % modulos
	shard := fmt.Sprint(bucket, "-", fmt.Sprint(modulos))

	_master, err := etcache.Get(path.Join(etcdKey, shard), false, false)
	if err != nil {
		return nil, err
	}
	master := _master.Node.Value

	httpClient := &http.Client{}
	// `Do` will complain if r.RequestURI is set so we unset it
	r.RequestURI = ""
	r.URL, err = url.Parse(master)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed request (%s) to %s.", resp.Status, r.URL.String())
	}
	return resp.Body, nil
}

func RouteHttp(w http.ResponseWriter, r *http.Request, etcdKey string, modulos uint64) {
	reader, err := Route(r, etcdKey, modulos)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
}

type multiReadCloser struct {
	readers []io.ReadCloser
}

func (mr *multiReadCloser) Read(p []byte) (n int, err error) {
	for len(mr.readers) > 0 {
		n, err = mr.readers[0].Read(p)
		if n > 0 || err != io.EOF {
			if err == io.EOF {
				// Don't return EOF yet. There may be more bytes
				// in the remaining readers.
				err = nil
			}
			return
		}
		err = mr.readers[0].Close()
		if err != nil {
			return
		}
		mr.readers = mr.readers[1:]
	}
	return 0, io.EOF
}

func (mr *multiReadCloser) Close() error {
	for len(mr.readers) > 0 {
		err := mr.readers[0].Close()
		if err != nil {
			return err
		}
		mr.readers = mr.readers[1:]
	}
	return nil
}

// MultiReadCloser returns a ReaderCloser that's the logical concatenation of
// the provided input readers. They're read sequentially.  Once all inputs have
// returned EOF, Read will return EOF. If any of the readers return a non-nil,
// non-EOF error, Read will return that error.  MultiReadCloser closes all of
// the input readers when it is closed. It also closes readers when they finish.
func MultiReadCloser(readers ...io.ReadCloser) io.ReadCloser {
	r := make([]io.ReadCloser, len(readers))
	copy(r, readers)
	return &multiReadCloser{r}
}

// Multicast enables the Ogre Magi to rapidly cast his spells, giving them
// greater potency.
func Multicast(r *http.Request, etcdKey string) (io.ReadCloser, error) {
	_endpoints, err := etcache.Get(etcdKey, false, true)
	if err != nil {
		return nil, err
	}
	endpoints := _endpoints.Node.Nodes

	var body []byte
	if r.ContentLength != 0 {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}

	var readers []io.ReadCloser
	for i, node := range endpoints {
		httpClient := &http.Client{}
		// `Do` will complain if r.RequestURI is set so we unset it
		r.RequestURI = ""
		r.URL, err = url.Parse(node.Value)
		if err != nil {
			return nil, err
		}

		if r.ContentLength != 0 {
			r.Body = ioutil.NopCloser(bytes.NewReader(body))
		}

		resp, err := httpClient.Do(r)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("Failed request (%s) to %s.", resp.Status, r.URL.String())
		}
		// paths with * are multigets so we want all of the responses
		if i == 0 || strings.Contains(r.URL.Path, "*") {
			readers = append(readers, resp.Body)
		}
	}

	return MultiReadCloser(readers...), nil
}

func MulticastHttp(w http.ResponseWriter, r *http.Request, etcdKey string) {
	reader, err := Multicast(r, etcdKey)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	defer reader.Close()

	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
}
