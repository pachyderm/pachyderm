package obj

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
)

// TestUserAgentSuffix verifies that an Amazon session built by this package
// appends a "pachyderm/<version>" token to the outgoing User-Agent without
// dropping the SDK-built value.
func TestUserAgentSuffix(t *testing.T) {
	sess, err := amazonSession(context.Background(), &ObjectStoreURL{
		Scheme: "s3",
		Bucket: "test-bucket",
		Params: "region=us-east-1",
	})
	if err != nil {
		t.Fatalf("amazonSession: %v", err)
	}

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
	if !strings.Contains(ua, s3UserAgentProduct+"/"+userAgentVersion()) {
		t.Fatalf("User-Agent missing pachyderm token, got: %q", ua)
	}
	// The handler must append, not replace: the SDK base token must survive.
	if !strings.Contains(strings.ToLower(ua), "aws-sdk-go") {
		t.Fatalf("User-Agent missing aws-sdk-go base token, got: %q", ua)
	}
}

func TestUserAgentVersionToken(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		want string
	}{
		{name: "release", in: "2.9.1", want: "2.9.1"},
		{name: "release candidate", in: "2.9.1rc1", want: "2.9.1rc1"},
		{name: "unstamped", in: "0.0.0", want: "dev"},
		{name: "space", in: "2.9.1 dirty", want: "dev"},
		{name: "newline", in: "2.9.1\r\nX-Test: injected", want: "dev"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := userAgentVersionToken(test.in); got != test.want {
				t.Fatalf("userAgentVersionToken(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}
