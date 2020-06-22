package testutil

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	cachedEnterpriseCodeOnce sync.Once
	cachedEnterpriseCode     string
	cachedEnterpriseSkipped  bool
)

// GetTestEnterpriseCode gets a Pachyderm Enterprise activation code from a
// private S3 bucket, and provides it to tests that use Pachyderm Enterprise
// features
func GetTestEnterpriseCode(t testing.TB) string {
	cachedEnterpriseCodeOnce.Do(func() {
		// Get test enterprise code from s3. The Pachyderm Enterprise test activation
		// token is stored in
		// s3://pachyderm-engineering/test_enterprise_activation_code.txt
		s3client := s3.New(session.Must(session.NewSessionWithOptions(
			session.Options{
				Config: aws.Config{
					Region: aws.String("us-west-1"), // contains s3://pachyderm-engineering
				},
			})))
		var s3Resp *s3.GetObjectOutput
		if err := backoff.Retry(func() error {
			var err error
			s3Resp, err = s3client.GetObject(&s3.GetObjectInput{
				Bucket: aws.String("pachyderm-engineering"),
				Key:    aws.String("test_enterprise_activation_code.txt"),
			})
			return err
		}, backoff.NewTestingBackOff()); err != nil {
			if strings.Contains(err.Error(), "NoCredentialProviders") {
				cachedEnterpriseSkipped = true
				return
			}
			// tests can't run without credentials -- just crash
			panic(fmt.Sprintf("cannot get test enterprise token from s3: %v", err))
		}
		buf := &bytes.Buffer{}
		if _, err := buf.ReadFrom(s3Resp.Body); err != nil {
			panic(fmt.Sprintf("cannot copy test enterprise token to buffer: %v", err))
		}
		cachedEnterpriseCode = buf.String()
	})
	if cachedEnterpriseSkipped {
		t.Skip("skipping test since credentials aren't available.")
	}
	return cachedEnterpriseCode
}
