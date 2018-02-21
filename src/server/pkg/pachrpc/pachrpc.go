package pachrpc

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client"
)

var (
	// pachClient is the "template" client that contains the original GRPC
	// connection and that other clients returned by this library are based on. It
	// has no ctx and therefore no auth credentials or cancellation
	pachClient *client.APIClient

	// pachClientMu guards pachClient
	pachClientOnce sync.Once

	// pachClientReady is closed once pachClient has been initialized
	pachClientReady = make(chan struct{})
)

// InitPachRPC initializes the PachRPC library. This creates the template
// pachClient used by future calls to GetPachClient and dials a GRPC connection
// to 'address'.
//
// This should be called by packages that receive incoming ctxs (which must be
// transformed to pachClients to propagate e.g. auth info). For example,
// {pfs,pps}.NewAPIServer should call this, as they receive incoming RPCs with
// attached ctxs, but newDriver, which only receives internal calls, should
// not (its caller--pps or any tests--would do that).
func InitPachRPC(address string) {
	pachClientOnce.Do(func() {
		if address == "" {
			panic("cannot initialize pach client with empty address")
		}
		if pachClient != nil {
			return // already initialized
		}
		var err error
		pachClient, err = client.NewFromAddress(address)
		if err != nil {
			panic(fmt.Sprintf("pps failed to initialize pach client: %v", err))
		}
		close(pachClientReady)
	})
}

// GetPachClient returns a pachd client with the same authentication credentials
// and cancellation as 'ctx' (ensuring that auth credentials are propagated
// through downstream RPCs). Internal Pachyderm calls accept clients returned
// by this call.
func GetPachClient(ctx context.Context) *client.APIClient {
	<-pachClientReady // wait until InitPachRPC is finished
	return pachClient.WithCtx(ctx)
}
