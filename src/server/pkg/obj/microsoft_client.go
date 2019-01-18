package obj

import (
	"io"

	"github.com/Azure/azure-sdk-for-go/storage"
)

const (
	maxBlockSize = 4 * 1024 * 1024 // 4MB (according to: https://docs.microsoft.com/en-us/rest/api/storageservices/understanding-block-blobs--append-blobs--and-page-blobs#about-append-blobs)
)

type microsoftClient struct {
	blobClient storage.BlobStorageClient
	container  string
}

func newMicrosoftClient(container string, accountName string, accountKey string) (*microsoftClient, error) {
	client, err := storage.NewBasicClient(
		accountName,
		accountKey,
	)
	if err != nil {
		return nil, err
	}

	return &microsoftClient{
		blobClient: client.GetBlobService(),
		container:  container,
	}, nil
}

func (c *microsoftClient) Writer(name string) (io.WriteCloser, error) {
	writer, err := newMicrosoftWriter(c, name)
	if err != nil {
		return nil, err
	}
	return newBackoffWriteCloser(c, writer), nil
}

func (c *microsoftClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	byteRange := byteRange(offset, size)
	var reader io.ReadCloser
	var err error
	if byteRange == "" {
		reader, err = c.blobClient.GetBlob(c.container, name)
	} else {
		reader, err = c.blobClient.GetBlobRange(c.container, name, byteRange, nil)
	}

	if err != nil {
		return nil, err
	}
	return newBackoffReadCloser(c, reader), nil
}

func (c *microsoftClient) Delete(name string) error {
	return c.blobClient.DeleteBlob(c.container, name, nil)
}

func (c *microsoftClient) Walk(name string, fn func(name string) error) error {
	// See Azure docs for what `marker` does:
	// https://docs.microsoft.com/en-us/rest/api/storageservices/List-Blobs?redirectedfrom=MSDN
	var marker string
	for {
		blobList, err := c.blobClient.ListBlobs(c.container, storage.ListBlobsParameters{
			Prefix: name,
			Marker: marker,
		})
		if err != nil {
			return err
		}

		for _, file := range blobList.Blobs {
			if err := fn(file.Name); err != nil {
				return err
			}
		}

		// NextMarker is empty when all results have been returned
		if blobList.NextMarker == "" {
			break
		}
		marker = blobList.NextMarker
	}
	return nil
}

func (c *microsoftClient) Exists(name string) bool {
	exists, _ := c.blobClient.BlobExists(c.container, name)
	return exists
}

func (c *microsoftClient) IsRetryable(err error) (ret bool) {
	microsoftErr, ok := err.(storage.AzureStorageServiceError)
	if !ok {
		return false
	}
	return microsoftErr.StatusCode >= 500
}

func (c *microsoftClient) IsNotExist(err error) bool {
	microsoftErr, ok := err.(storage.AzureStorageServiceError)
	if !ok {
		return false
	}
	return microsoftErr.StatusCode == 404
}

func (c *microsoftClient) IsIgnorable(err error) bool {
	return false
}

type microsoftWriter struct {
	container  string
	blob       string
	blobClient storage.BlobStorageClient
}

func newMicrosoftWriter(client *microsoftClient, name string) (*microsoftWriter, error) {
	// create container
	if _, err := client.blobClient.CreateContainerIfNotExists(client.container, storage.ContainerAccessTypePrivate); err != nil {
		return nil, err
	}

	// create blob
	if err := client.blobClient.PutAppendBlob(client.container, name, nil); err != nil {
		return nil, err
	}

	return &microsoftWriter{
		container:  client.container,
		blob:       name,
		blobClient: client.blobClient,
	}, nil
}

func (w *microsoftWriter) Write(b []byte) (int, error) {
	nBytes := 0
	for i := 0; i < len(b); i += maxBlockSize {
		var chunk []byte
		if i+maxBlockSize >= len(b) {
			chunk = b[i:]
		} else {
			chunk = b[i : i+maxBlockSize]
		}
		if err := w.blobClient.AppendBlock(w.container, w.blob, chunk, nil); err != nil {
			return nBytes, err
		}
		nBytes += len(chunk)
	}
	return nBytes, nil
}

func (w *microsoftWriter) Close() error {
	return nil
}
