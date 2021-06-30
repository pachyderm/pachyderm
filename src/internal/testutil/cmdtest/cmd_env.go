package cmdtest

import (
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

type TestEnv struct {
	MockClient           *MockClient
	MockEnterpriseClient *MockClient
	Stdin                io.Reader
	Stdout               io.Writer
	Stderr               io.Writer
}

func NewTestEnv(t *testing.T, stdin io.Reader, stdout io.Writer, stderr io.Writer) *TestEnv {
	return &TestEnv{
		MockClient:           NewMockClient(t),
		MockEnterpriseClient: NewMockClient(t),
		Stdin:                stdin,
		Stdout:               stdout,
		Stderr:               stderr,
	}
}

func (env *TestEnv) Client(string, ...client.Option) *client.APIClient {
	return env.MockClient.APIClient
}

func (env *TestEnv) EnterpriseClient(string, ...client.Option) *client.APIClient {
	return env.MockEnterpriseClient.APIClient
}

func (env *TestEnv) In() io.Reader {
	return env.Stdin
}

func (env *TestEnv) Out() io.Writer {
	return env.Stdout
}

func (env *TestEnv) Err() io.Writer {
	return env.Stderr
}

func (env *TestEnv) Close() error {
	return nil
}
