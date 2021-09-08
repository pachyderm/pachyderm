package testutil

import (
	"os"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
)

var (
	pachClient *client.APIClient
	pachErr    error
	clientOnce sync.Once
)

// GetPachClient gets a pachyderm client for use in tests. It works for tests
// running both inside and outside the cluster. Note that multiple calls to
// GetPachClient will return the same instance.
func GetPachClient(t testing.TB) *client.APIClient {
	// to use opentracing within a test, PACH_TRACE and JAEGER_ENDPOINT must be set
	_, useOpenTrace := os.LookupEnv(tracing.ShortTraceEnvVar)
	if useOpenTrace {
		tracing.InstallJaegerTracerFromEnv()
		if _, ok := os.LookupEnv(extended.TraceDurationEnvVar); !ok {
			err := os.Setenv(extended.TraceDurationEnvVar, "5m")
			if err != nil {
				t.Fatalf("error getting Pachyderm client: %s", err.Error())
			}
		}
	}

	clientOnce.Do(func() {
		if _, ok := os.LookupEnv("PACHD_PORT_1650_TCP_ADDR"); ok {
			pachClient, pachErr = client.NewInCluster()
		} else {
			pachClient, pachErr = client.NewForTest()
		}
	})
	if pachErr != nil {
		t.Fatalf("error getting Pachyderm client: %s", pachErr.Error())
	}

	if useOpenTrace {
		ctx, err := extended.EmbedAnyDuration(pachClient.Ctx())
		if err != nil {
			t.Fatalf("Could not embed duration in Pachyderm client: %s", err.Error())
		}
		pachClient = pachClient.WithCtx(ctx)
	}
	return pachClient
}
