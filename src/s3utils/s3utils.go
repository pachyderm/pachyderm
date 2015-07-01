package s3utils

import (
	"bytes"
	"io"
	"log"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

const (
	// AWS says that parts must be at least 5MB, it's unclear if that means 5 *
	// 10^6 or 5 2^10 so we went with the larger.
	minPart = 5242880      // 5MB
	maxPart = minPart * 10 // 50MB
)

// An s3 input looks like: s3://bucket/dir
// Where dir can be a path

// getBucket extracts the bucket from an s3 input
func GetBucket(input string) (string, error) {
	return strings.Split(strings.TrimPrefix(input, "s3://"), "/")[0], nil
}

// getPath extracts the path from an s3 input
func GetPath(input string) (string, error) {
	return path.Join(strings.Split(strings.TrimPrefix(input, "s3://"), "/")[1:]...), nil
}

func NewBucket(uri string) (*s3.Bucket, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	client := s3.New(auth, aws.USWest)
	bucket, err := GetBucket(uri)
	if err != nil {
		return nil, err
	}

	return client.Bucket(bucket), nil
}

// PutMulti is like a smart bucket.Put in that it will automatically do a
// multiput if the input reader has enough data that it makes sense to do so.
func PutMulti(bucket *s3.Bucket, path string, r io.Reader, contType string, perm s3.ACL) error {
	// A pointer to a Multi transaction, if this is non nil it means we're
	// putting the data in parts.
	var multi *s3.Multi = nil
	var parts []s3.Part
	for i := 1; ; i++ {
		data := make([]byte, maxPart)
		n, err := io.ReadAtLeast(r, data, minPart)
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			// EOF means that r was empty
			// ErrUnexpectedEOF means that r had less than minPart bytes in it
			// Note that we can hit these errors after having read several 5MB
			// chunks from r already
			log.Print(err)
			return err
		}
		if n >= minPart && multi == nil {
			// We have enough data to do a MultiPut but we haven't started on yet.
			// That means it's time to start one
			multi, err = bucket.Multi(path, contType, perm)
			if err != nil {
				log.Print(err)
				return err
			}
		}
		// Now we upload the actual data
		if multi == nil {
			// We're not doing a multi transaction,
			return bucket.Put(path, data[0:n], contType, perm)
		} else {
			part, err := multi.PutPart(i, bytes.NewReader(data[0:n]))
			if err != nil {
				log.Print(err)
				return err
			}
			parts = append(parts, part)
		}
		if n < minPart {
			// That means this was the last batch of data, time to break out of
			// this loop
			break
		}
	}
	if multi != nil {
		if err := multi.Complete(parts); err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

// Files calls `cont` on each file found at `uri` starting at marker.
// Pass `marker=""` to start from the beginning.
// Returns the marker that should be passed to pick-up where this call left off.
func ForEachFile(uri, marker string, cont func(file string, modtime time.Time) error) error {
	bucket, err := NewBucket(uri)
	if err != nil {
		return err
	}
	inPath, err := GetPath(uri)
	if err != nil {
		return err
	}
	nextMarker := marker
	for {
		lr, err := bucket.List(inPath, "", nextMarker, 0)
		if err != nil {
			return err
		}
		for _, key := range lr.Contents {
			modtime, err := time.Parse(time.RFC3339, key.LastModified)
			if err != nil {
				log.Print(err)
				return err
			}
			err = cont(key.Key, modtime)
			if err != nil {
				return err
			}
			nextMarker = key.Key
		}
		if !lr.IsTruncated {
			// We've exhausted the output
			break
		}
	}
	return nil
}
