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
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"go.uber.org/zap"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const namespace = "fuzz-cluster-1"
const useSucccessValWeight = 1

var (
	run               = flag.Bool("testgen.run", true, "be default, don't run workflow data generation") // DNJ TODO - just hide in test harness? set back to false at least
	alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
	successInputs     = map[string][]interface{}{}
)

var protosUnderTest = map[protoreflect.FileDescriptor][]string{ // TODO - make own rangeRPC function from rpc_fuzz_test
	pfs.File_pfs_pfs_proto: {
		pfs.API_CreateRepo_FullMethodName,
		pfs.API_CreateProject_FullMethodName,
		pfs.API_DeleteRepo_FullMethodName,
		pfs.API_DeleteProject_FullMethodName,
		pfs.API_InspectProjectV2_FullMethodName,
		pfs.API_InspectRepo_FullMethodName,
	},
}

type genData struct { // DNJ TODO - awful name
	r     *rand.Rand
	field protoreflect.FieldDescriptor
}

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
			"prxoy.resources.requests.memory": "2Gi",
			"prxoy.resources.limits.memory":   "2Gi",
			"proxy.replicas":                  "5", // proxy overload_manager oom kills incoming requests to avoid taking down the cluster, which is smart, but not what we want here.
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

func inputGenerator(ctx context.Context, r *rand.Rand, msgDesc protoreflect.MessageDescriptor, potentialInputs map[string][]interface{}) *dynamicpb.Message { // DNJ TODO - generics maybe?
	msg := dynamicpb.NewMessage(msgDesc)
	// DNJ TODO Inject and save successful field values
	defaultGenerators := map[protoreflect.Kind]generator{
		protoreflect.BoolKind:    func(gen genData) interface{} { return gen.r.Float32() < .5 }, // DNJ TODO - Generics and/or pass context for error?
		protoreflect.StringKind:  func(gen genData) interface{} { return randString(gen.r, 0, 5) },
		protoreflect.MessageKind: func(gen genData) interface{} { return inputGenerator(ctx, gen.r, gen.field.Message(), potentialInputs) },
	}
	for i := 0; i < msgDesc.Fields().Len(); i++ { // DNJ TODO fuzz each type and recurse
		field := msgDesc.Fields().Get(i)
		var fieldVal interface{}
		log.Info(ctx, "DNJ TODO setting field", zap.String("field", string(field.FullName())), zap.String("msg", string(msgDesc.FullName()))) // DNJ TODO repeated fields
		if successVals, ok := potentialInputs[string(field.FullName())]; ok &&                                                                // DNJ TODO figure out input naming
			r.Float32() < useSucccessValWeight {
			fieldVal = successVals[r.Intn(len(successVals))]
			log.Info(ctx, "DNJ TODO setting field success", zap.Any("field", field), zap.Any("fieldVal", fieldVal))
		} else {
			gen := genData{r: r, field: field}
			fieldVal = defaultGenerators[field.Kind()](gen)
		}
		msg.Set(field, protoreflect.ValueOf(fieldVal))
	}
	return msg
}
func storeSuccessVals(ctx context.Context, msg *dynamicpb.Message, out map[string][]interface{}) {
	msg.Range(func(field protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		log.Info(ctx, "DNJ TODO adding field", zap.Any("field", field), zap.Any("field val", val.Interface()))
		fieldName := string(field.FullName())
		fieldVal := val.Interface()
		if vals, ok := out[fieldName]; ok {
			vals = append(vals, fieldVal) // DNJ TODO - test
		} else {
			out[fieldName] = []interface{}{fieldVal} // DNJ TODO repeated fields
		}
		if field.Kind() == protoreflect.MessageKind { // handle sub messages
			subMsg, ok := fieldVal.(*dynamicpb.Message)
			log.Info(ctx, "DNJ TODO sub msg", zap.Any("submsg", subMsg))
			if ok { //should always be ok because of the .Kind() check
				storeSuccessVals(ctx, subMsg, out)
			}
		}
		return true // always continue since there's no error case
	})
	// t.Logf("DNJ TODO %#v", out)
}

func fuzzGrpc(ctx context.Context, c *client.APIClient, seed int64) { // DNJ TODO - can and/or should this be public?
	r := rand.New(rand.NewSource(seed)) // DNJ TODO - accept intrface that Rand stisifies to allow custom value sourcing? move successvals logic there! needs Float32 and Intn right now
	rangeRPCsList(protosUnderTest, func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) {
		input := inputGenerator(ctx, r, md.Input(), successInputs)
		reply := dynamicpb.NewMessage(md.Output())
		fullName := "/" + path.Join(string(sd.FullName()), string(md.Name())) // DNJ TODO move to rpc fuzz test and make function for this.
		if err := c.ClientConn().Invoke(ctx, fullName, input, reply); err != nil {
			log.Info(ctx, "Invoke grpc method err", zap.String("method name", fullName), zap.Error(err))
		} else {
			log.Info(ctx, "successful rpc call, saving fields for re-use", zap.String("method name", fullName))
			storeSuccessVals(ctx, input, successInputs)
			storeSuccessVals(ctx, reply, successInputs)
		}
		// t.Logf("Reply: %#v", reply)
	})
}

// DNJ TODO context and logger?
func FuzzGrpcWorkflow(f *testing.F) { // DNJ TODO - better name
	if !*run {
		f.Skip()
	}
	l := log.InitBatchLogger("/tmp/FuzzGrpc.log") //fmt.Sprintf("/tmp/FuzzGrpcLog-%d", os.Getpid()))
	ctx := pctx.Background("FuzzGrpcWorkflow")
	c := getSharedCluster(f)
	// for i := 0; i < 100; i++ {
	// 	fuzzGrpc(f, ctx, c, time.Now().Unix())
	// 	time.Sleep(time.Millisecond * 1000)
	// }
	f.Fuzz(func(t *testing.T, seed int64) {
		fuzzGrpc(ctx, c, seed)
		l(nil)
	})

}
