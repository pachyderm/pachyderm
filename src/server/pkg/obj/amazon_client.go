package obj

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

type amazonClient struct {
	bucket                 string
	cloudfrontDistribution string
	cloudfrontURLSigner    *sign.URLSigner
	s3                     *s3.S3
	uploader               *s3manager.Uploader
}

func newAmazonClient(bucket string, cloudfrontDistribution string, id string, secret string, token string, region string) (*amazonClient, error) {
	session := session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(id, secret, token),
		Region:      aws.String(region),
	})
	var signer *sign.URLSigner
	cloudfrontDistribution = strings.TrimSpace(cloudfrontDistribution)
	if cloudfrontDistribution != "" {
		rawCloudfrontPrivateKey, err := ioutil.ReadFile("/amazon-secret/cloudfrontPrivateKey")
		if err != nil {
			return nil, err
		}

		cloudfrontKeyPairID, err := ioutil.ReadFile("/amazon-secret/cloudfrontKeyPairId")
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(bytes.TrimSpace(rawCloudfrontPrivateKey))
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return nil, fmt.Errorf("block undefined or wrong type: type is (%v) should be (RSA PRIVATE KEY)", block.Type)
		}
		cloudfrontPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		signer = sign.NewURLSigner(string(cloudfrontKeyPairID), cloudfrontPrivateKey)
		log.Infof("Using cloudfront security credentials - keypair ID (%v) - to sign cloudfront URLs", string(cloudfrontKeyPairID))
	}
	return &amazonClient{
		bucket:                 bucket,
		cloudfrontDistribution: cloudfrontDistribution,
		cloudfrontURLSigner:    signer,
		s3:                     s3.New(session),
		uploader:               s3manager.NewUploader(session),
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
	var reader io.ReadCloser
	if c.cloudfrontDistribution != "" {
		var resp *http.Response
		var connErr error
		url := fmt.Sprintf("http://%v.cloudfront.net/%v", c.cloudfrontDistribution, name)

		if c.cloudfrontURLSigner != nil {
			signedURL, err := c.cloudfrontURLSigner.Sign(url, time.Now().Add(1*time.Hour))
			if err != nil {
				return nil, err
			}
			url = strings.TrimSpace(signedURL)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Range", byteRange)

		backoff.RetryNotify(func() error {
			resp, connErr = http.DefaultClient.Do(req)
			if connErr != nil && isNetRetryable(connErr) {
				return connErr
			}
			return nil
		}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) {
			log.Infof("Error connecting to (%v); retrying in %s: %#v", url, d, err)
		})
		if connErr != nil {
			return nil, connErr
		}
		if resp.StatusCode >= 300 {
			// Cloudfront returns 200s, and 206s as success codes
			return nil, fmt.Errorf("cloudfront returned HTTP error code %v for url %v", resp.Status, url)
		}
		reader = resp.Body
	} else {
		getObjectOutput, err := c.s3.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(name),
			Range:  aws.String(byteRange),
		})
		if err != nil {
			return nil, err
		}
		reader = getObjectOutput.Body
	}
	return newBackoffReadCloser(c, reader), nil
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

func (c *amazonClient) isRetryable(err error) (retVal bool) {
	if strings.Contains(err.Error(), "unexpected EOF") {
		return true
	}

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
	if c.cloudfrontDistribution != "" {
		// cloudfront returns forbidden error for nonexisting data
		if strings.Contains(err.Error(), "error code 403") {
			return true
		}
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
