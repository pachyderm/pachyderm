package obj

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/storagegateway"
)

type amazonClient struct {
	bucket   string
	s3       *s3.S3
	uploader *s3manager.Uploader
}

func newAmazonClient(bucket string, id string, secret string, token string, region string) (*amazonClient, error) {
	session := session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(id, secret, token),
		Region:      aws.String(region),
	})
	return &amazonClient{
		bucket:   bucket,
		s3:       s3.New(session),
		uploader: s3manager.NewUploader(session),
	}, nil
}

func (c *amazonClient) Writer(name string) (io.WriteCloser, error) {
	return newBackoffWriteCloser(c, newWriter(c, name)), nil
}

func (c *amazonClient) Walk(name string, fn func(name string) error) error {
	var fnErr error
	if err := c.s3.ListObjectsPages(
		&s3.ListObjectsInput{
			Bucket: aws.String(c.bucket),
			Prefix: aws.String(name),
		},
		func(listObjectsOutput *s3.ListObjectsOutput, lastPage bool) bool {
			for _, object := range listObjectsOutput.Contents {
				if err := fn(*object.Key); err != nil {
					fnErr = err
					return false
				}
			}
			return true
		},
	); err != nil {
		return err
	}
	return fnErr
}

func (c *amazonClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	byteRange := byteRange(offset, size)
	if byteRange != "" {
		byteRange = fmt.Sprintf("bytes=%s", byteRange)
	}

	// being a little fast and loose here ... not putting the GET within a retry loop ... but will be an interesting test for now
	url := fmt.Sprintf("http://d2z5sy3mh7px6z.cloudfront.net/%v", name)
	resp, err := http.Get(url)
	//	defer resp.Body.Close()
	fmt.Printf("got resp, err: %v,\n\n%v\n", resp, err)
	if err != nil {
		fmt.Printf("Got http error: %v\n", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		fmt.Printf("HTTP error code %v", resp.StatusCode)
		return nil, fmt.Errorf("HTTP error code %v", resp.StatusCode)
	}
	n, err := io.CopyN(ioutil.Discard, resp.Body, int64(offset))
	fmt.Printf("slurped n bytes: %v off of get file %v to accomodate offset\n", n, url)
	if err != nil {
		return nil, err
	}
	//return newBackoffReadCloser(c, getObjectOutput.Body), nil
	return resp.Body, nil
}

func (c *amazonClient) Delete(name string) error {
	_, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
	})
	return err
}

func (c *amazonClient) Exists(name string) bool {
	_, err := c.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
	})
	return err == nil
}

func (c *amazonClient) isRetryable(err error) bool {
	awsErr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	for _, c := range []string{
		storagegateway.ErrorCodeServiceUnavailable,
		storagegateway.ErrorCodeInternalError,
		storagegateway.ErrorCodeGatewayInternalError,
	} {
		if c == awsErr.Code() {
			return true
		}
	}
	return false
}

func (c *amazonClient) IsIgnorable(err error) bool {
	return false
}

func (c *amazonClient) IsNotExist(err error) bool {
	fmt.Printf("IsNotExist? error: %v\n", err)
	// cloudfront returns forbidden error for nonexisting data
	if strings.Contains(err.Error(), "error code 403") {
		fmt.Printf("its a 403, dne")
		return true
	}
	if strings.Contains(err.Error(), "error code 404") {
		fmt.Printf("its a 404, dne")
		return true
	}
	awsErr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	if awsErr.Code() == "NoSuchKey" {
		return true
	}
	return false
}

type amazonWriter struct {
	errChan chan error
	pipe    *io.PipeWriter
}

func newWriter(client *amazonClient, name string) *amazonWriter {
	reader, writer := io.Pipe()
	w := &amazonWriter{
		errChan: make(chan error),
		pipe:    writer,
	}
	go func() {
		_, err := client.uploader.Upload(&s3manager.UploadInput{
			Body:            reader,
			Bucket:          aws.String(client.bucket),
			Key:             aws.String(name),
			ContentEncoding: aws.String("application/octet-stream"),
		})
		w.errChan <- err
	}()
	return w
}

func (w *amazonWriter) Write(p []byte) (int, error) {
	return w.pipe.Write(p)
}

func (w *amazonWriter) Close() error {
	if err := w.pipe.Close(); err != nil {
		return err
	}
	return <-w.errChan
}
