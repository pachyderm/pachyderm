package obj

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
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
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	log "github.com/sirupsen/logrus"
)

type amazonClient struct {
	bucket                 string
	cloudfrontDistribution string
	cloudfrontURLSigner    *sign.URLSigner
	s3                     *s3.S3
	uploader               *s3manager.Uploader
	advancedConfig         *AmazonAdvancedConfiguration
}

// AmazonCreds are options that are applicable specifically to Pachd's
// credentials in an AWS deployment
type AmazonCreds struct {
	// Direct credentials. Only applicable if Pachyderm is given its own permanent
	// AWS credentials
	ID     string // Access Key ID
	Secret string // Secret Access Key
	Token  string // Access token (if using temporary security credentials
}

func parseLogOptions(optstring string) *aws.LogLevelType {
	if optstring == "" {
		return nil
	}
	toLevel := map[string]aws.LogLevelType{
		"debug":           aws.LogDebug,
		"signing":         aws.LogDebugWithSigning,
		"httpbody":        aws.LogDebugWithHTTPBody,
		"requestretries":  aws.LogDebugWithRequestRetries,
		"requesterrors":   aws.LogDebugWithRequestErrors,
		"eventstreambody": aws.LogDebugWithEventStreamBody,
		"all": aws.LogDebugWithSigning |
			aws.LogDebugWithHTTPBody |
			aws.LogDebugWithRequestRetries |
			aws.LogDebugWithRequestErrors |
			aws.LogDebugWithEventStreamBody,
	}
	var result aws.LogLevelType
	opts := strings.Split(optstring, ",")
	for _, optStr := range opts {
		result |= toLevel[strings.ToLower(optStr)]
	}
	var msg bytes.Buffer
	// build log message separately, as the log flags have overlapping definitions
	msg.WriteString("using S3 logging flags: ")
	for _, optStr := range []string{"Debug", "Signing", "HTTPBody", "RequestRetries", "RequestErrors", "EventStreamBody"} {
		optBits := toLevel[strings.ToLower(optStr)]
		if (result & optBits) == optBits {
			msg.WriteString(optStr)
			msg.WriteString(",")
		}
	}
	log.Infof(msg.String())
	return &result
}

func newAmazonClient(region, bucket string, creds *AmazonCreds, cloudfrontDistribution string, endpoint string, advancedConfig *AmazonAdvancedConfiguration) (*amazonClient, error) {
	// set up aws config, including credentials (If creds.ID not set then this will use the EC2 metadata service)
	timeout, err := time.ParseDuration(advancedConfig.Timeout)
	if err != nil {
		return nil, errors.EnsureStack(err)
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
		LogLevel:   parseLogOptions(advancedConfig.LogOptions),
		Logger:     aws.NewDefaultLogger(),
	}
	if creds.ID != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(creds.ID, creds.Secret, creds.Token)
	}
	// Set custom endpoint for a custom deployment.
	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	// Create new session using awsConfig
	session, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.EnsureStack(err)
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
			return nil, errors.EnsureStack(err)
		}
		awsClient.cloudfrontURLSigner = sign.NewURLSigner(cloudfrontKeyPairID, cloudfrontPrivateKey)
		log.Infof("Using cloudfront security credentials - keypair ID (%v) - to sign cloudfront URLs", string(cloudfrontKeyPairID))
	}
	return awsClient, nil
}

func (c *amazonClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	_, err := c.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		ACL:             aws.String(c.advancedConfig.UploadACL),
		Body:            r,
		Bucket:          aws.String(c.bucket),
		Key:             aws.String(name),
		ContentEncoding: aws.String("application/octet-stream"),
	})
	return err
}

func (c *amazonClient) Walk(ctx context.Context, name string, fn func(name string) error) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	var fnErr error
	var prefix = &name
	if err := c.s3.ListObjectsPagesWithContext(ctx,
		&s3.ListObjectsInput{
			Bucket: aws.String(c.bucket),
			Prefix: prefix,
		},
		func(listObjectsOutput *s3.ListObjectsOutput, lastPage bool) bool {
			for _, object := range listObjectsOutput.Contents {
				key := *object.Key
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

func (c *amazonClient) Get(ctx context.Context, name string, w io.Writer) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	var reader io.ReadCloser
	if c.cloudfrontDistribution != "" {
		var resp *http.Response
		var connErr error
		url := fmt.Sprintf("http://%v.cloudfront.net/%v", c.cloudfrontDistribution, name)

		if c.cloudfrontURLSigner != nil {
			signedURL, err := c.cloudfrontURLSigner.Sign(url, time.Now().Add(1*time.Hour))
			if err != nil {
				return err
			}
			url = strings.TrimSpace(signedURL)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

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
			return connErr
		}
		if resp.StatusCode >= 300 {
			// Cloudfront returns 200s, and 206s as success codes
			return errors.Errorf("cloudfront returned HTTP error code %v for url %v", resp.Status, url)
		}
		reader = resp.Body
	} else {
		objIn := &s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(name),
		}
		getObjectOutput, err := c.s3.GetObject(objIn)
		if err != nil {
			return err
		}
		reader = getObjectOutput.Body
	}
	defer func() {
		if err := reader.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err := io.Copy(w, reader)
	return err
}

func (c *amazonClient) Delete(ctx context.Context, name string) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	_, err := c.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
	})
	return err
}

func (c *amazonClient) Exists(ctx context.Context, name string) (bool, error) {
	_, err := c.s3.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
	})
	tracing.TagAnySpan(ctx, "err", err)
	if err != nil {
		err = c.transformError(err, name)
		if pacherr.IsNotExist(err) {
			err = nil
		}
		return false, err
	}
	return true, nil
}

func (c *amazonClient) transformError(err error, objectPath string) error {
	const minWait = 250 * time.Millisecond
	if err == nil {
		return nil
	}
	if c.cloudfrontDistribution != "" {
		// cloudfront returns forbidden error for nonexisting data
		if strings.Contains(err.Error(), "error code 403") {
			return pacherr.NewNotExist(c.bucket, objectPath)
		}
	}
	if strings.Contains(err.Error(), "unexpected EOF") {
		return pacherr.WrapTransient(err, minWait)
	}
	if strings.Contains(err.Error(), "Not Found") {
		return pacherr.NewNotExist(c.bucket, objectPath)
	}
	var awsErr awserr.Error
	if !errors.As(err, &awsErr) {
		return err
	}
	if strings.Contains(awsErr.Message(), "SlowDown:") {
		return pacherr.WrapTransient(err, minWait)
	}
	switch awsErr.Code() {
	case s3.ErrCodeNoSuchKey:
		return pacherr.NewNotExist(c.bucket, objectPath)
	case storagegateway.ErrorCodeServiceUnavailable,
		storagegateway.ErrorCodeInternalError,
		storagegateway.ErrorCodeGatewayInternalError:
		return pacherr.WrapTransient(err, minWait)
	}
	return err
}

func isNetRetryable(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Temporary()
}
