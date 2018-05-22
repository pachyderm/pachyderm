package obj

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type microsoftClient struct {
	blobClient storage.BlobStorageClient
	container  *storage.Container
}

func newMicrosoftClient(container string, accountName string, accountKey string) (*microsoftClient, error) {
	client, err := storage.NewBasicClient(
		accountName,
		accountKey,
	)
	if err != nil {
		return nil, err
	}

	blobClient := client.GetBlobService()
	return &microsoftClient{
		blobClient: blobClient,
		container:  blobClient.GetContainerReference(container),
	}, nil
}

func (c *microsoftClient) blob(name string) *storage.Blob {
	return c.container.GetBlobReference(name)
}

func (c *microsoftClient) Writer(name string) (io.WriteCloser, error) {
	writer, err := c.newMicrosoftWriter(name)
	if err != nil {
		return nil, err
	}
	return newBackoffWriteCloser(c, writer), nil
}

func (c *microsoftClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error
	if offset == 0 && size == 0 {
		reader, err = c.blob(name).Get(nil)
	} else {
		end := uint64(0)
		if size != 0 {
			end = offset + size
		}
		reader, err = c.blob(name).GetRange(&storage.GetBlobRangeOptions{
			Range: &storage.BlobRange{
				Start: offset,
				End:   end,
			},
		})
	}

	if err != nil {
		return nil, err
	}
	return newBackoffReadCloser(c, reader), nil
}

func (c *microsoftClient) Delete(name string) error {
	return c.blob(name).Delete(nil)
}

func (c *microsoftClient) Copy(src, dst string) error {
	return c.blob(dst).Copy(c.blob(src).GetURL(), nil)
}

func (c *microsoftClient) Walk(name string, fn func(name string) error) error {
	// See Azure docs for what `marker` does:
	// https://docs.microsoft.com/en-us/rest/api/storageservices/List-Blobs?redirectedfrom=MSDN
	var marker string
	for {
		blobList, err := c.container.ListBlobs(storage.ListBlobsParameters{
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
	exists, _ := c.blob(name).Exists()
	return exists
}

func (c *microsoftClient) Stat(name string) (*ObjInfo, error) {
	// populate the blob's properties
	blob := c.blob(name)
	if err := blob.GetProperties(nil); err != nil {
		return nil, err
	}
	return &ObjInfo{
		Name: name,
		Size: blob.Properties.ContentLength,
	}, nil
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
	blobClient storage.BlobStorageClient
	blob       *storage.Blob
}

func (c *microsoftClient) newMicrosoftWriter(name string) (*microsoftWriter, error) {
	// create container
	_, err := c.container.CreateIfNotExists(&storage.CreateContainerOptions{Access: storage.ContainerAccessTypePrivate})
	if err != nil {
		return nil, err
	}

	// create blob
	blob := c.blob(name)
	err = blob.CreateBlockBlob(nil)
	if err != nil {
		return nil, err
	}

	return &microsoftWriter{
		blobClient: c.blobClient,
		blob:       blob,
	}, nil
}

func (w *microsoftWriter) Write(b []byte) (int, error) {
	blockList, err := w.blob.GetBlockList(storage.BlockListTypeAll, nil)
	if err != nil {
		return 0, err
	}

	blocksLen := len(blockList.CommittedBlocks)
	amendList := []storage.Block{}
	for _, v := range blockList.CommittedBlocks {
		amendList = append(amendList, storage.Block{v.Name, storage.BlockStatusCommitted})
	}

	inputSourceReader := bytes.NewReader(b)
	chunk := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(chunk)
	for {
		n, err := inputSourceReader.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		blockID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%011d\n", blocksLen)))
		data := chunk[:n]
		err = w.blob.PutBlock(blockID, data, nil)
		if err != nil {
			return 0, fmt.Errorf("BlobStorageClient.PutBlock: %v", err)
		}
		// add current uncommitted block to temporary block list
		amendList = append(amendList, storage.Block{blockID, storage.BlockStatusUncommitted})
		blocksLen++
	}

	// update block list to blob committed block list.
	err = w.blob.PutBlockList(amendList, nil)
	if err != nil {
		return 0, fmt.Errorf("BlobStorageClient.PutBlockList: %v", err)
	}
	return len(b), nil
}

func (w *microsoftWriter) Close() error {
	return nil
}
