//go:build k8s

package workflowfuzz

import (
	"context"
	"flag"
	"math/rand"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const namespace = "fuzz-cluster-1"
const useSucccessfulValWeight = .8

var (
	run               = flag.Bool("testgen.run", true, "be default, don't run workflow data generation") // DNJ TODO - just hide in test harness? set back to false at least
	alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
	successfulInputs  = map[string][]interface{}{}
)

var protosUnderTest = map[protoreflect.FileDescriptor][]string{ // TODO - make own rangeRPC function from rpc_fuzz_test
	pfs.File_pfs_pfs_proto: {pfs.API_CreateRepo_FullMethodName, pfs.API_CreateProject_FullMethodName},
}

type genData struct { // DNJ TODO - awful name
	r     *rand.Rand
	field protoreflect.FieldDescriptor
}

// type grpcArrayWriter struct {
// 	ByteCount   int
// 	GrpcMethods []string
// }

//	func (w *grpcArrayWriter) Write(bytes []byte) (int, error) {
//		if len(bytes) > 0 {
//			if w.GrpcMethods == nil {
//				w.GrpcMethods = make([]string, 0)
//			}
//			w.GrpcMethods = append(w.GrpcMethods, string(bytes))
//			w.ByteCount += len(bytes)
//		}
//		return w.ByteCount, nil
//	}
//
// DNJ TODO - rpc fuzz test
func rangeRPCsList(protos map[protoreflect.FileDescriptor][]string, f func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor)) {
	for fd, protoDescriptors := range protos {
		svcMethods := map[string][]string{}
		for _, name := range protoDescriptors {
			split := strings.Split(name, "/") // DNJ TODO - what is up with the / separator?
			svcName := split[1]               // 0 is empty
			methodName := split[2]
			if arr, ok := svcMethods[svcName]; !ok {
				svcMethods[svcName] = []string{methodName}
			} else {
				svcMethods[svcName] = append(arr, methodName)
			}
		}
		svcs := fd.Services()
		for si := 0; si < svcs.Len(); si++ {
			sd := svcs.Get(si)
			if methodNames, ok := svcMethods[string(sd.FullName())]; ok {
				methods := sd.Methods()
				for mi := 0; mi < methods.Len(); mi++ {
					md := methods.Get(mi)
					if slices.Contains(methodNames, string(md.Name())) {
						f(fd, sd, md)
					}
				}
			}
		}
	}
}

func getSharedCluster(f *testing.F) *client.APIClient {
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
	return pachClient
}

func randString(r *rand.Rand, minlength int, maxLength int) string {
	var characters strings.Builder
	length := r.Intn(maxLength) + minlength
	for i := 0; i < length; i++ {
		characters.WriteByte(byte(r.Intn(127)))
	}
	return characters.String()
}

type generator func(genData) interface{}

func inputGenerator(r *rand.Rand, msgDesc protoreflect.MessageDescriptor) protoreflect.ProtoMessage { // DNJ TODO - generics maybe?
	msg := dynamicpb.NewMessage(msgDesc)
	// DNJ TODO Inject and save successful field values
	defaultGenerators := map[protoreflect.Kind]generator{
		protoreflect.BoolKind:    func(gen genData) interface{} { return gen.r.Float32() < .5 }, // DNJ TODO - Generics and/or pass context for error?
		protoreflect.StringKind:  func(gen genData) interface{} { return randString(gen.r, 0, 5) },
		protoreflect.MessageKind: func(gen genData) interface{} { return inputGenerator(gen.r, gen.field.Message()) },
	}
	for i := 0; i < msgDesc.Fields().Len(); i++ { // DNJ TODO fuzz each type and recurse
		field := msgDesc.Fields().Get(i)
		var fieldVal interface{}
		if successVals, ok := successfulInputs[string(field.FullName())]; ok && r.Float32() < useSucccessfulValWeight {
			fieldVal = successVals[r.Intn(len(successVals))]
		} else {
			gen := genData{r: r, field: field}
			fieldVal = defaultGenerators[field.Kind()](gen)
		}
		msg.Set(field, protoreflect.ValueOf(fieldVal))
	}
	return msg
}

// DNJ TODO context and logger?
func FuzzGrpcWorkflow(f *testing.F) { // DNJ TODO - better name
	if !*run {
		f.Skip()
	}
	ctx := pctx.Background("FuzzGrpcWorkflow")
	// log.InitPachctlLogger()
	// ctx := pctx.Background("workflow-fuzz")
	client := getSharedCluster(f)
	// grpcs := &grpcArrayWriter{} // DNJ TODO - needed? do we just list from file descriptors? We need the "testable" whitelist. Should be protosUnderTest
	// err := cmds.GrpcFromAddress(fmt.Sprintf("%s:%d", grpcAddress.Host, grpcAddress.Port)).
	// 	Run(pctx.TestContext(f),
	// 		&pachctl.Config{},
	// 		grpcs,
	// 		[]string{},
	// 	)
	// require.NoError(f, err)
	// f.Logf("grpc method: %v len: %d", grpcs.GrpcMethods, grpcs.ByteCount)
	// DNJ TODO - magic to select repo rpc
	// selectedMethod := "pfs_v2.API.CreateRepo"

	// for i := 0; i < 50; i++ {
	// 	f.Add(randutil.UniqueString("repo"))
	// }
	// r := rand.New(rand.NewSource(time.Now().Unix()))

	f.Fuzz(func(t *testing.T, seed int64) {
		// if !alphaNumericRegex.MatchString(name) || name == "" { // DNJ TODO remove err
		// 	t.Skip()
		// }
		r := rand.New(rand.NewSource(seed))
		// f.Logf("fd : %v -sd: %v -md: %v msg: %v", fd.FullName(), sd.Name(), md.Input().Fields(), inputGenerator(r, md.Input()))
		rangeRPCsList(protosUnderTest, func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) {
			input := inputGenerator(r, md.Input())
			reply := dynamicpb.NewMessage(md.Output())
			fullName := "/" + path.Join(string(sd.FullName()), string(md.Name())) // DNJ TODO move to rpc fuzz test and make function for this.
			if err := client.ClientConn().Invoke(ctx, fullName, input, reply); err != nil {
				t.Logf("Invoke grpc method err %s", err.Error())
			}
			t.Logf("Reply: %#v", reply)
		})

		// request := &pfs.CreateRepoRequest{
		// 	Repo: &pfs.Repo{
		// 		Name: string(name),
		// 		Type: "user",
		// 		Project: &pfs.Project{
		// 			Name: "default",
		// 		},
		// 	},
		// 	Update: false,
		// }
		// requestBytes, err := json.Marshal(request)
		// require.NoError(t, err)
		// err = cmds.GrpcFromAddress(fmt.Sprintf("%s:%d", grpcAddress.Host, grpcAddress.Port)).
		// 	Run(pctx.TestContext(t),
		// 		&pachctl.Config{},
		// 		responseBytes,
		// 		[]string{strings.ReplaceAll(strings.TrimLeft(pfs.API_CreateRepo_FullMethodName, "/"), "/", "."), //"pfs_v2.API.CreateRepo",
		// 			string(requestBytes),
		// 		})
		// if err != nil &&
		// 	!strings.Contains(err.Error(), "only alphanumeric characters") &&
		// 	!strings.Contains(err.Error(), "already exists") &&
		// 	!strings.Contains(err.Error(), "invalid escape code") &&
		// 	!strings.Contains(err.Error(), "invalid character") {
		// 	// log.Info(ctx, "Err in RPC", zap.Error(err))
		// 	t.Errorf("Why fail? %v", err)
		// }
		// // require.NoError(t, err)
		// t.Logf("RESP: %s", string(responseBytes.Bytes()))
	})

}
