package cmdtest

import (
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

type TestEnv struct {
	MockClient           *MockClient
	MockEnterpriseClient *MockClient
	In                   io.Reader
	Out                  io.Writer
	Err                  io.Writer
}

func NewTestEnv(t *testing.T, stdin io.Reader, stdout io.Writer, stderr io.Writer) *TestEnv {
	return &TestEnv{
		MockClient:           NewMockClient(t),
		MockEnterpriseClient: NewMockClient(t),
		In:                   stdin,
		Out:                  stdout,
		Err:                  stderr,
	}
}

func (env *TestEnv) Client(string, ...client.Option) *client.APIClient {
	return env.MockClient.APIClient
}

func (env *TestEnv) EnterpriseClient(string, ...client.Option) *client.APIClient {
	return env.MockEnterpriseClient.APIClient
}

func (env *TestEnv) Stdin() io.Reader {
	return env.In
}

func (env *TestEnv) Stdout() io.Writer {
	return env.Out
}

func (env *TestEnv) Stderr() io.Writer {
	return env.Err
}

func (env *TestEnv) Close() error {
	return nil
}
