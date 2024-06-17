package obj

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// ObjectStoreURL represents a parsed URL to an object in an object store.
type ObjectStoreURL struct {
	// The object store, e.g. s3, gcs, as...
	Scheme string
	// The "bucket" (in AWS parlance) or the "container" (in Azure parlance).
	Bucket string
	// The object itself.
	Object string
	// The query parameters if defined.
	Params string
}

func (s ObjectStoreURL) String() string {
	bucket := fmt.Sprintf("%s://%s/%s", s.Scheme, s.Bucket, s.Object)
	if s.Params != "" {
		bucket += "?" + s.Params
	}
	return bucket
}

func (s ObjectStoreURL) BucketString() string {
	bucket := fmt.Sprintf("%s://%s", s.Scheme, s.Bucket)
	if s.Params != "" {
		bucket += "?" + s.Params
	}
	return bucket
}

// ParseURL parses an URL into ObjectStoreURL.
func ParseURL(urlStr string) (*ObjectStoreURL, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing url %v", urlStr)
	}
	// TODO: remove once we remove support for non-existing schemes.
	if u.Scheme == "gcs" {
		u.Scheme = "gs"
	}
	var objStoreUrl *ObjectStoreURL
	switch u.Scheme {
	case "azblob", "gs", "file", "s3":
		objStoreUrl = &ObjectStoreURL{
			Scheme: u.Scheme,
			Bucket: u.Host,
			Object: strings.Trim(u.Path, "/"),
			Params: u.RawQuery,
		}
	// wasb is an hdfs protocol supported by azure
	// the format is wasb://<container_name>@<storage_account_name>.blob.core.windows.net/dir/file
	// another supported format is wasb://<container_name>@<storage_account_name>/dir/file
	case "wasb":
		objStoreUrl = &ObjectStoreURL{
			Scheme: "azblob",
			Bucket: u.User.Username(),
			Object: strings.Trim(u.Path, "/"),
			Params: u.RawQuery,
		}
	case "minio", "test-minio":
		parts := strings.SplitN(strings.Trim(u.Path, "/"), "/", 2)
		var key string
		if len(parts) == 2 {
			key = parts[1]
		}
		objStoreUrl = &ObjectStoreURL{
			Scheme: u.Scheme,
			Bucket: u.Host + "/" + parts[0],
			Object: key,
			Params: u.RawQuery,
		}
	default:
		return nil, errors.Errorf("unrecognized object store: %s", u.Scheme)
	}
	return objStoreUrl, nil
}
