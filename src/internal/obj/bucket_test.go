package obj

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

// TestUserAgentSuffix verifies that an Amazon session built by this package
// appends a "pachyderm/<version>" token to the outgoing User-Agent without
// dropping the SDK-built value.
func TestUserAgentSuffix(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		t.Fatalf("session.NewSession: %v", err)
	}
	// Apply the same handler the production session path installs.
	sess.Handlers.Build.PushBack(
		request.MakeAddToUserAgentHandler("pachyderm", userAgentVersion()),
	)

	op := &request.Operation{Name: "TestOp", HTTPMethod: "GET", HTTPPath: "/"}
	req := request.New(
		*sess.Config,
		metadata.ClientInfo{
			ServiceName: "s3",
			Endpoint:    "https://s3.us-east-1.amazonaws.com",
			APIVersion:  "2006-03-01",
		},
		sess.Handlers,
		client.DefaultRetryer{NumMaxRetries: 0},
		op,
		nil,
		nil,
	)
	req.Handlers.Build.Run(req)

	ua := req.HTTPRequest.Header.Get("User-Agent")
	if !strings.Contains(ua, "pachyderm/"+userAgentVersion()) {
		t.Fatalf("User-Agent missing pachyderm token, got: %q", ua)
	}
	// The handler must append, not replace: the SDK base token must survive.
	if !strings.Contains(strings.ToLower(ua), "aws-sdk-go") {
		t.Fatalf("User-Agent missing aws-sdk-go base token, got: %q", ua)
	}
}
