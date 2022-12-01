package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pachyderm/pachyderm/v2/src/client"
	uuid "github.com/satori/go.uuid"
	// "go.uber.org/ratelimit"
)

var (
	size = flag.String("size", "50MB", "size of the file to upload")
	kind = flag.String("kind", "grpc", "type of benchmark to perform")
	rate = flag.String("limit", "", "rate limit in bytes per second for the upload")
)

type R struct {
	b   byte
	Len uint64
	//Ratelimit ratelimit.Limiter
	lastRead time.Time
	EOFAt    time.Time
}

func (r *R) Read(p []byte) (int, error) {
	// if r.Ratelimit == nil {
	// 	r.Ratelimit = ratelimit.NewUnlimited()
	// }
	var n int
	for i := range p {
		if r.Len > 0 {
			//r.Ratelimit.Take()
			p[i] = r.b
			r.Len--
			n++
		} else {
			if n > 0 {
				fmt.Printf(".")
			} else {
				fmt.Printf("\n")
			}
			r.EOFAt = time.Now()
			return n, io.EOF
		}
	}
	r.b++
	fmt.Printf(".")
	r.lastRead = time.Now()
	return n, nil
}

var _ io.Reader = new(R)

func bench(f func(name string, r io.Reader, length uint64) error) error {
	// if *rate != "" {
	// 	if x, err := humanize.ParseBytes(*rate); err != nil {
	// 		log.Fatalf("parse rate: %v", err)
	// 	} else {
	// 		rl := ratelimit.New(int(x))
	// 		r.Ratelimit = rl
	// 	}
	// } else {
	// 	r.Ratelimit = ratelimit.NewUnlimited()
	// }
	for i := 0; i < 10; i++ {
		r := new(R)
		var length uint64
		if x, err := humanize.ParseBytes(*size); err != nil {
			log.Fatalf("parse size: %v", err)
		} else {
			r.Len = x
			length = x
		}

		n, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}
		name := n.String()
		start := time.Now()
		if err := f(name, r, length); err != nil {
			return err
		}
		log.Printf("total time: %v", time.Since(start).String())
		log.Printf(" = %s/s", humanize.Bytes(uint64(float64(length)/float64(time.Since(start).Seconds()))))
		log.Printf("flush time: %v", time.Since(r.EOFAt).String())
		log.Printf("read time: %v", time.Since(start)-time.Since(r.EOFAt))
	}
	return nil
}

func main() {
	flag.Parse()

	c, err := client.NewOnUserMachine("")
	// c, err := client.NewFromURI("grpc://localhost:9001")
	if err != nil {
		log.Fatal(err)
	}
	if err := c.CreateProjectRepo("default", "benchmark-upload"); err != nil {
		log.Printf("create repo: %v", err)
	}

	var benchErr error
	commit := client.NewProjectCommit("default", "benchmark-upload", "master", "")
	switch *kind {
	case "grpc":
		benchErr = bench(func(name string, r io.Reader, _ uint64) error {
			return c.PutFile(commit, name, r)
		})
	case "http":
		benchErr = bench(func(name string, r io.Reader, _ uint64) error {
			req, err := http.NewRequest("PUT", "http://localhost/upload/"+name, r)
			if err != nil {
				return err
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			b, err := httputil.DumpResponse(res, true)
			if err != nil {
				return err
			}
			log.Printf("%s", b)
			return nil
		})
	case "s3":
		mc, err := minio.New("localhost", &minio.Options{
			Creds: credentials.NewStaticV4(c.AuthToken(), c.AuthToken(), ""),
		})
		if err != nil {
			log.Fatalf("minio: %v", err)
		}
		benchErr = bench(func(name string, r io.Reader, length uint64) error {
			res, err := mc.PutObject(context.TODO(), "master.benchmark-upload", name, r, int64(length), minio.PutObjectOptions{})
			log.Printf("%#v", res)
			return err
		})
	default:
		log.Fatalf("unknown benchmark kind %v", *kind)
	}
	if benchErr != nil {
		log.Fatal(benchErr)
	}

	if err := c.Close(); err != nil {
		log.Fatalf("close: %v", err)
	}
}
