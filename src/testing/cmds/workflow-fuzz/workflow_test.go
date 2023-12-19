//go:build k8s

package workflowfuzz

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/misc/cmds"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const namespace = "fuzz-cluster-1"

var run = flag.Bool("testgen.run", false, "be default, don't run workflow data generation") // DNJ TODO - just hide in test harness?
var alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
var protosUnderTest = map[protoreflect.FileDescriptor][]string{ // TODO - make own rangeRPC function from rpc_fuzz_test
	pfs.File_pfs_pfs_proto: {pfs.API_CreateRepo_FullMethodName},
}

type grpcArrayWriter struct {
	ByteCount   int
	GrpcMethods []string
}

func (w *grpcArrayWriter) Write(bytes []byte) (int, error) {
	if len(bytes) > 0 {
		if w.GrpcMethods == nil {
			w.GrpcMethods = make([]string, 0)
		}
		w.GrpcMethods = append(w.GrpcMethods, string(bytes))
		w.ByteCount += len(bytes)
	}
	return w.ByteCount, nil
}

func getSharedCluster(f *testing.F) *grpcutil.PachdAddress {
	k := testutil.GetKubeClient(f)
	minikubetestenv.PutNamespace(f, namespace)
	pachClient := minikubetestenv.InstallRelease(f, context.Background(), namespace, k, &minikubetestenv.DeployOpts{
		CleanupAfter:       false,
		UseLeftoverCluster: true,
		ValueOverrides: map[string]string{
			"prxoy.resources.requests.memory": "2Gi", // DNJ TODO -- proxy overload_manager oom kills incoming requests to avoid taking down the cluster, which is smart, but not what we want.
			"prxoy.resources.limits.memory":   "2Gi",
			"proxy.replicas":                  "3",
		},
	})
	return pachClient.GetAddress()
}

func randString(r *rand.Rand, maxLength int) string {
	var characters strings.Builder
	length := r.Intn(maxLength)
	for i := 0; i < length; i++ {
		characters.WriteByte(byte(r.Intn(127)))
	}
	return characters.String()
}

// DNJ TODO context and logger?
func FuzzGrpcWorkflow(f *testing.F) { // DNJ TODO - better name
	if !*run {
		f.Skip()
	}
	// log.InitPachctlLogger()
	// ctx := pctx.Background("workflow-fuzz")
	grpcAddress := getSharedCluster(f)
	grpcs := &grpcArrayWriter{} // DNJ TODO - needed? do we just list from file descriptors? We need the "testable" whitelist. Should be protosUnderTest
	err := cmds.GrpcFromAddress(fmt.Sprintf("%s:%d", grpcAddress.Host, grpcAddress.Port)).
		Run(pctx.TestContext(f),
			&pachctl.Config{},
			grpcs,
			[]string{},
		)
	require.NoError(f, err)
	f.Logf("grpc method: %v len: %d", grpcs.GrpcMethods, grpcs.ByteCount)
	// DNJ TODO - magic to select repo rpc
	// selectedMethod := "pfs_v2.API.CreateRepo"

	// for i := 0; i < 50; i++ {
	// 	f.Add(randutil.UniqueString("repo"))
	// }
	f.Fuzz(func(t *testing.T, seed int64) {
		// if !alphaNumericRegex.MatchString(name) || name == "" { // DNJ TODO remove err
		// 	t.Skip()
		// }
		r := rand.New(rand.NewSource(seed))
		name := randString(r, 10)
		t.Logf("input: %s", string(name))
		responseBytes := &bytes.Buffer{}
		request := &pfs.CreateRepoRequest{
			Repo: &pfs.Repo{
				Name: string(name),
				Type: "user",
				Project: &pfs.Project{
					Name: "default",
				},
			},
			Update: false,
		}
		requestBytes, err := json.Marshal(request)
		require.NoError(t, err)
		err = cmds.GrpcFromAddress(fmt.Sprintf("%s:%d", grpcAddress.Host, grpcAddress.Port)).
			Run(pctx.TestContext(t),
				&pachctl.Config{},
				responseBytes,
				[]string{strings.ReplaceAll(strings.TrimLeft(pfs.API_CreateRepo_FullMethodName, "/"), "/", "."), //"pfs_v2.API.CreateRepo",
					string(requestBytes),
				})
		if err != nil &&
			!strings.Contains(err.Error(), "only alphanumeric characters") &&
			!strings.Contains(err.Error(), "already exists") &&
			!strings.Contains(err.Error(), "invalid escape code") &&
			!strings.Contains(err.Error(), "invalid character") {
			// log.Info(ctx, "Err in RPC", zap.Error(err))
			t.Errorf("Why fail? %v", err)
		}
		// require.NoError(t, err)
		t.Logf("RESP: %s", string(responseBytes.Bytes()))
	})

}
