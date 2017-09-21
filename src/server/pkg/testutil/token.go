package testutil

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	cachedEnterpriseCodeMut sync.Mutex
	cachedEnterpriseCode    string
)

// GetTestEnterpriseCode gets a Pachyderm Enterprise activation code from a
// private S3 bucket, and provides it to tests that use Pachyderm Enterprise
// features
func GetTestEnterpriseCode() string {
	cachedEnterpriseCodeMut.Lock()
	defer cachedEnterpriseCodeMut.Unlock()

	if cachedEnterpriseCode != "" {
		return cachedEnterpriseCode
	}

	// Get test enterprise code from s3. The Pachyderm Enterprise test activation
	// token is stored in
	// s3://pachyderm-engineering/test_enterprise_activation_code.txt
	s3client := s3.New(session.Must(session.NewSessionWithOptions(
		session.Options{
			Config: aws.Config{
				Region: aws.String("us-west-1"), // contains s3://pachyderm-engineering
			},
		})))
	// S3 client requires string pointers. Go does not allow programs to take the
	// address of constants, so we must create strings here
	output, err := s3client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("pachyderm-engineering"),
		Key:    aws.String("test_enterprise_activation_code.txt"),
	})
	if err != nil {
		// tests can't run without credentials -- just crash
		panic(fmt.Sprintf("cannot get test enterprise token from s3: %v", err))
	}
	buf := &bytes.Buffer{}
	if _, err = buf.ReadFrom(output.Body); err != nil {
		panic(fmt.Sprintf("cannot copy test enterprise token to buffer: %v", err))
	}
	cachedEnterpriseCode = buf.String()
	return cachedEnterpriseCode
}
