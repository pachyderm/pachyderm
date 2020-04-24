package obj

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	log "github.com/sirupsen/logrus"

	vault "github.com/hashicorp/vault/api"
)

const oneDayInSeconds = 60 * 60 * 24
const twoDaysInSeconds = 60 * 60 * 48

type amazonClient struct {
	bucket                 string
	cloudfrontDistribution string
	cloudfrontURLSigner    *sign.URLSigner
	s3                     *s3.S3
	uploader               *s3manager.Uploader
	advancedConfig         *AmazonAdvancedConfiguration
}

type vaultCredentialsProvider struct {
	vaultClient *vault.Client // client used to retrieve S3 creds from vault
	vaultRole   string        // get vault creds from: /aws/creds/<vaultRole>

	// ID, duration, and last renew time of the vault lease that governs the most
	// recent AWS secret issued to this pachd instance (and a mutex protecting
	// them)
	leaseMu        sync.Mutex
	leaseID        string
	leaseLastRenew time.Time
	leaseDuration  time.Duration
}

// updateLease extracts the duration of the lease governing 'secret' (an AWS
// secret). IIUC, because the AWS backend issues dynamic secrets, there is no
// tokens associated with them, and vaultSecret.TokenTTL can be ignored
func (v *vaultCredentialsProvider) updateLease(secret *vault.Secret) {
	v.leaseMu.Lock()
	defer v.leaseMu.Unlock()
	v.leaseID = secret.LeaseID
	v.leaseLastRenew = time.Now()
	v.leaseDuration = time.Duration(secret.LeaseDuration) * time.Second
}

func (v *vaultCredentialsProvider) getLeaseDuration() time.Duration {
	v.leaseMu.Lock()
	defer v.leaseMu.Unlock()
	return v.leaseDuration
}

// Retrieve returns nil if it successfully retrieved the value.  Error is
// returned if the value were not obtainable, or empty.
func (v *vaultCredentialsProvider) Retrieve() (credentials.Value, error) {
	var emptyCreds, result credentials.Value // result

	// retrieve AWS creds from vault
	vaultSecret, err := v.vaultClient.Logical().Read(path.Join("aws", "creds", v.vaultRole))
	if err != nil {
		return emptyCreds, errors.Wrapf(err, "could not retrieve creds from vault")
	}
	accessKeyIface, accessKeyOk := vaultSecret.Data["access_key"]
	awsSecretIface, awsSecretOk := vaultSecret.Data["secret_key"]
	if !accessKeyOk || !awsSecretOk {
		return emptyCreds, errors.Errorf("aws creds not present in vault response")
	}

	// Convert access key & secret in response to strings
	result.AccessKeyID, accessKeyOk = accessKeyIface.(string)
	result.SecretAccessKey, awsSecretOk = awsSecretIface.(string)
	if !accessKeyOk || !awsSecretOk {
		return emptyCreds, errors.Errorf("aws creds in vault response were not both strings (%T and %T)", accessKeyIface, awsSecretIface)
	}

	// update the lease values in 'v', and spawn a goroutine to renew the lease
	v.updateLease(vaultSecret)
	go func() {
		for {
			// renew at half the lease duration or one day, whichever is greater
			// (lease must expire eventually)
			renewInterval := v.getLeaseDuration()
			if renewInterval.Seconds() < oneDayInSeconds {
				renewInterval = oneDayInSeconds * time.Second
			}

			// Wait until 'renewInterval' has elapsed, then renew the lease
			time.Sleep(renewInterval)
			backoff.RetryNotify(func() error {
				// every two days, renew the lease for this node's AWS credentials
				vaultSecret, err := v.vaultClient.Sys().Renew(v.leaseID, twoDaysInSeconds)
				if err != nil {
					return err
				}
				v.updateLease(vaultSecret)
				return nil
			}, backoff.NewExponentialBackOff(), func(err error, _ time.Duration) error {
				log.Errorf("could not renew vault lease: %v", err)
				return nil
			})
		}
	}()

	// Per https://www.vaultproject.io/docs/secrets/aws/index.html#usage, wait
	// until token is usable
	time.Sleep(10 * time.Second)
	return result, nil
}

// IsExpired returns if the credentials are no longer valid, and need to be
// retrieved.
func (v *vaultCredentialsProvider) IsExpired() bool {
	v.leaseMu.Lock()
	defer v.leaseMu.Unlock()
	return time.Now().After(v.leaseLastRenew.Add(v.leaseDuration))
}

// AmazonCreds are options that are applicable specifically to Pachd's
// credentials in an AWS deployment
type AmazonCreds struct {
	// Direct credentials. Only applicable if Pachyderm is given its own permanent
	// AWS credentials
	ID     string // Access Key ID
	Secret string // Secret Access Key
	Token  string // Access token (if using temporary security credentials

	// Vault options (if getting AWS credentials from Vault)
	VaultAddress string // normally addresses come from env, but don't have vault service name
	VaultRole    string
	VaultToken   string
}

func newAmazonClient(region, bucket string, creds *AmazonCreds, cloudfrontDistribution string, endpoint string, advancedConfig *AmazonAdvancedConfiguration) (*amazonClient, error) {
	// set up aws config, including credentials (if neither creds.ID nor
	// creds.VaultAddress are set, then this will use the EC2 metadata service
	timeout, err := time.ParseDuration(advancedConfig.Timeout)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{Timeout: timeout}
	// If NoVerifySSL is true, then configure the transport to skip ssl verification (enables self-signed certificates).
	if advancedConfig.NoVerifySSL {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	}
	awsConfig := &aws.Config{
		Region:     aws.String(region),
		MaxRetries: aws.Int(advancedConfig.Retries),
		HTTPClient: httpClient,
		DisableSSL: aws.Bool(advancedConfig.DisableSSL),
	}
	if creds.ID != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(creds.ID, creds.Secret, creds.Token)
	} else if creds.VaultAddress != "" {
		vaultClient, err := vault.NewClient(&vault.Config{
			Address: creds.VaultAddress,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error creating vault client")
		}
		vaultClient.SetToken(creds.VaultToken)
		awsConfig.Credentials = credentials.NewCredentials(&vaultCredentialsProvider{
			vaultClient: vaultClient,
			vaultRole:   creds.VaultRole,
		})
	}
	// Set custom endpoint for a custom deployment.
	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	// Create new session using awsConfig
	session, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}
	awsClient := &amazonClient{
		bucket: bucket,
		s3:     s3.New(session),
		uploader: s3manager.NewUploader(session, func(u *s3manager.Uploader) {
			u.PartSize = advancedConfig.PartSize
			u.MaxUploadParts = advancedConfig.MaxUploadParts
		}),
		advancedConfig: advancedConfig,
	}

	// Set awsClient.cloudfrontURLSigner and cloudfrontDistribution (if Pachd is
	// using cloudfront)
	awsClient.cloudfrontDistribution = strings.TrimSpace(cloudfrontDistribution)
	if cloudfrontDistribution != "" {
		rawCloudfrontPrivateKey, err := readSecretFile("/cloudfrontPrivateKey")
		if err != nil {
			return nil, err
		}
		cloudfrontKeyPairID, err := readSecretFile("/cloudfrontKeyPairId")
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(bytes.TrimSpace([]byte(rawCloudfrontPrivateKey)))
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return nil, errors.Errorf("block undefined or wrong type: type is (%v) should be (RSA PRIVATE KEY)", block.Type)
		}
		cloudfrontPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		awsClient.cloudfrontURLSigner = sign.NewURLSigner(cloudfrontKeyPairID, cloudfrontPrivateKey)
		log.Infof("Using cloudfront security credentials - keypair ID (%v) - to sign cloudfront URLs", string(cloudfrontKeyPairID))
	}
	return awsClient, nil
}

func (c *amazonClient) Writer(ctx context.Context, name string) (io.WriteCloser, error) {
	if c.advancedConfig.Reverse {
		name = reverse(name)
	}
	return newBackoffWriteCloser(ctx, c, newWriter(ctx, c, name)), nil
}

func (c *amazonClient) Walk(_ context.Context, name string, fn func(name string) error) error {
	var fnErr error
	var prefix *string

	if c.advancedConfig.Reverse {
		prefix = nil
	} else {
		prefix = &name
	}

	if err := c.s3.ListObjectsPages(
		&s3.ListObjectsInput{
			Bucket: aws.String(c.bucket),
			Prefix: prefix,
		},
		func(listObjectsOutput *s3.ListObjectsOutput, lastPage bool) bool {
			for _, object := range listObjectsOutput.Contents {
				key := *object.Key
				if c.advancedConfig.Reverse {
					key = reverse(key)
				}
				if strings.HasPrefix(key, name) {
					if err := fn(key); err != nil {
						fnErr = err
						return false
					}
				}
			}
			return true
		},
	); err != nil {
		return err
	}
	return fnErr
}

func (c *amazonClient) Reader(ctx context.Context, name string, offset uint64, size uint64) (io.ReadCloser, error) {
	if c.advancedConfig.Reverse {
		name = reverse(name)
	}
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

		backoff.RetryNotify(func() (retErr error) {
			span, _ := tracing.AddSpanToAnyExisting(ctx, "/Amazon.Cloudfront/Get")
			defer func() {
				tracing.FinishAnySpan(span, "err", retErr)
			}()
			resp, connErr = http.DefaultClient.Do(req)
			if connErr != nil && isNetRetryable(connErr) {
				return connErr
			}
			return nil
		}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
			log.Infof("Error connecting to (%v); retrying in %s: %#v", url, d, err)
			return nil
		})
		if connErr != nil {
			return nil, connErr
		}
		if resp.StatusCode >= 300 {
			// Cloudfront returns 200s, and 206s as success codes
			return nil, errors.Errorf("cloudfront returned HTTP error code %v for url %v", resp.Status, url)
		}
		reader = resp.Body
	} else {
		objIn := &s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(name),
		}
		if byteRange != "" {
			objIn.Range = aws.String(byteRange)
		}
		getObjectOutput, err := c.s3.GetObject(objIn)
		if err != nil {
			return nil, err
		}
		reader = getObjectOutput.Body
	}
	return newCheckedReadCloser(size, newBackoffReadCloser(ctx, c, reader)), nil
}

func (c *amazonClient) Delete(_ context.Context, name string) error {
	if c.advancedConfig.Reverse {
		name = reverse(name)
	}
	_, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
	})
	return err
}

func (c *amazonClient) Exists(ctx context.Context, name string) bool {
	if c.advancedConfig.Reverse {
		name = reverse(name)
	}
	_, err := c.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
	})
	tracing.TagAnySpan(ctx, "err", err)
	return err == nil
}

func (c *amazonClient) IsRetryable(err error) (retVal bool) {
	if strings.Contains(err.Error(), "unexpected EOF") {
		return true
	}
	if strings.Contains(err.Error(), "SlowDown:") {
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
	ctx     context.Context
	errChan chan error
	pipe    *io.PipeWriter
}

func newWriter(ctx context.Context, client *amazonClient, name string) *amazonWriter {
	reader, writer := io.Pipe()
	w := &amazonWriter{
		ctx:     ctx,
		errChan: make(chan error),
		pipe:    writer,
	}
	go func() {
		_, err := client.uploader.Upload(&s3manager.UploadInput{
			ACL:             aws.String(client.advancedConfig.UploadACL),
			Body:            reader,
			Bucket:          aws.String(client.bucket),
			Key:             aws.String(name),
			ContentEncoding: aws.String("application/octet-stream"),
		})
		if err != nil {
			reader.CloseWithError(err)
		}
		w.errChan <- err
	}()
	return w
}

func (w *amazonWriter) Write(p []byte) (retN int, retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(w.ctx, "/Amazon.Writer/Write")
	defer tracing.FinishAnySpan(span, "bytes", retN, "err", retErr)
	return w.pipe.Write(p)
}

func (w *amazonWriter) Close() (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(w.ctx, "/Amazon.Writer/Close")
	defer tracing.FinishAnySpan(span, "err", retErr)
	if err := w.pipe.Close(); err != nil {
		return err
	}
	return <-w.errChan
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
