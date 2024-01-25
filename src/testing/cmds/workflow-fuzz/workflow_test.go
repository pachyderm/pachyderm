//go:build k8s

package workflowfuzz

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	namespace            = "fuzz-cluster-1"
	useSucccessValWeight = .75
	maxRepeatedCount     = 5
	// usNilWeight          = .2
)

var (
	runDefault      = flag.Bool("gen.default", true, "use the default data generators") // DNJ TODO - just hide in test harness? set back to false at least
	successInputs   = map[string][]protoreflect.Value{}
	successInputsMu = sync.Mutex{}
)

var protosUnderTest = map[protoreflect.FileDescriptor][]string{ // DNJ TODO - make own rangeRPC function from rpc_fuzz_test? at least move list of apis to separate input file(s)
	pfs.File_pfs_pfs_proto: {
		pfs.API_CreateRepo_FullMethodName,
		pfs.API_CreateProject_FullMethodName,
		// pfs.API_DeleteRepo_FullMethodName,
		// pfs.API_DeleteProject_FullMethodName, // DNJ TODO - weighting?
		// pfs.API_InspectProjectV2_FullMethodName,
		// pfs.API_InspectRepo_FullMethodName,
		pfs.API_ListProject_FullMethodName,
		pfs.API_ListRepo_FullMethodName,
		// pfs.API_ListBranch_FullMethodName,
	},
	pps.File_pps_pps_proto: {
		pps.API_CreatePipeline_FullMethodName,
		// pps.API_InspectPipeline_FullMethodName,
		pps.API_ListPipeline_FullMethodName,
	},
}

type GeneratorIndex struct {
	KindGenerators map[protoreflect.Kind]Generator
	NameGenerators map[string]Generator
}

type generatorSource struct { // DNJ TODO - should this hold successInputs?
	ctx   context.Context
	r     *rand.Rand
	field protoreflect.FieldDescriptor
	// msg      *dynamicpb.Message
	hasInputLevel *int
	ancestry      []protoreflect.FieldDescriptor
	index         GeneratorIndex
}

type Generator func(generatorSource) protoreflect.Value

func subinput(genSource generatorSource) protoreflect.Value {
	if len(genSource.ancestry) > genSource.r.Intn(15) {
		return protoreflect.ValueOf(nil) // no multi-inputs
	} else {
		if len(genSource.ancestry) > 1 {
			log.Info(genSource.ctx, "DNJ TODO Input generator going deeper.",
				zap.Int("level", *genSource.hasInputLevel),
				zap.Any("ancestry", len(genSource.ancestry)),
				zap.Any("ancestor", genSource.ancestry[len(genSource.ancestry)-2].FullName()),
			)
		}
		*genSource.hasInputLevel = len(genSource.ancestry)
		return protoreflect.ValueOf(inputGenerator(genSource.ctx, genSource.r, genSource.field.Message(), successInputs, genSource)) // DNJ TODO - this depth lineage is too confusing
	}
}

func pfsInput(genSource generatorSource) protoreflect.Value {
	//DNJ TODO - how to know we are at the end of the input chain to avoid termination weight?
	if *genSource.hasInputLevel >= len(genSource.ancestry) || .35 > genSource.r.Float32() { // DNJ TODO magic number - inputTerminationWeight?
		return protoreflect.ValueOf(nil)
	} else {
		log.Info(genSource.ctx, "DNJ TODO Input generator stopping with pfs input.",
			zap.Int("level", *genSource.hasInputLevel),
			zap.Any("ancestry", len(genSource.ancestry)),
			zap.Any("ancestor", genSource.ancestry[len(genSource.ancestry)-2].FullName()),
		)
		*genSource.hasInputLevel = len(genSource.ancestry) // Only one input per level allowed
		return protoreflect.ValueOf(inputGenerator(genSource.ctx, genSource.r, genSource.field.Message(), successInputs, genSource))
	}
}

func noValue(genSource generatorSource) protoreflect.Value {
	return protoreflect.ValueOf(nil)
}

func constantVal(value interface{}) Generator {
	return func(genSource generatorSource) protoreflect.Value { return protoreflect.ValueOf(value) }
}

func randInt64Positive(genSource generatorSource) protoreflect.Value { // DNJ  TODO parametrize and return function?
	return protoreflect.ValueOf(genSource.r.Int63n(2000))
}

func randNanos(genSource generatorSource) protoreflect.Value {
	return protoreflect.ValueOf(genSource.r.Int31n(2000))
}

func randJson(genSource generatorSource) protoreflect.Value {
	return protoreflect.ValueOf("") // DNJ TODO - random json generator
}

func existingString(field string) Generator {
	return func(genSource generatorSource) protoreflect.Value {
		successInputsMu.Lock()
		defer successInputsMu.Unlock()
		if vals, ok := successInputs[field]; ok { // repo name doesn't match the field for pipeline inputs
			return vals[genSource.r.Intn(len(vals))]
		} else {
			return protoreflect.ValueOf(randString(genSource.r, 0, 20))
		}
	}
}

func randReproccessSpec(genSource generatorSource) protoreflect.Value {
	if genSource.r.Float32() < .5 {
		return protoreflect.ValueOf("until_success")
	}
	return protoreflect.ValueOf("every_job")
}

func randTolerationOperator(genSource generatorSource) protoreflect.Value {
	enum := genSource.field.Enum().Values()
	selected := genSource.r.Intn(enum.Len()-1) + 1 // 0 is invalid for create pipeline
	return protoreflect.ValueOf(enum.Get(selected).Number())
}

func randName(genSource generatorSource) protoreflect.Value {
	validChars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_")
	var characters strings.Builder
	length := genSource.r.Intn(20) + 1
	for i := 0; i < length; i++ {
		characters.WriteRune(validChars[genSource.r.Intn(len(validChars))])
	}
	return protoreflect.ValueOf(characters.String())
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

func getSharedCluster(t testing.TB) *client.APIClient {
	k := testutil.GetKubeClient(t)
	minikubetestenv.PutNamespace(t, namespace)
	pachClient := minikubetestenv.InstallRelease(t, context.Background(), namespace, k, &minikubetestenv.DeployOpts{
		CleanupAfter:       false,
		UseLeftoverCluster: true,
		ValueOverrides: map[string]string{
			"prxoy.resources.requests.memory": "1Gi",
			"prxoy.resources.limits.memory":   "1Gi",
			"proxy.replicas":                  "10", // proxy overload_manager oom kills incoming requests to avoid taking down the cluster, which is smart, but not what we want here.
			"console.enabled":                 "true",
			"pachd.enterpriseLicenseKey":      os.Getenv("ENT_ACT_CODE"),
			"pachd.activateAuth":              "false",
		},
	})
	return pachClient
}

func randString(r *rand.Rand, minlength int, maxLength int) string {
	var characters strings.Builder
	length := r.Intn(maxLength) + minlength
	for i := 0; i < length; i++ {
		characters.WriteByte(byte(32 + r.Intn(94))) // printable ascii
	}
	return characters.String()
}

func inputGenerator(ctx context.Context, r *rand.Rand, msgDesc protoreflect.MessageDescriptor, potentialInputs map[string][]protoreflect.Value, prevSource generatorSource) *dynamicpb.Message { // DNJ TODO - generics maybe?
	msg := dynamicpb.NewMessage(msgDesc)
	// DNJ TODO type enum?
	defaultGenerators := GeneratorIndex{ // DNJ TODO - full custom inputs, how to handle for specific fields? Maybe do my own hash with field name or have a field name map of generators that is checlked first?
		KindGenerators: map[protoreflect.Kind]Generator{
			protoreflect.BoolKind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(gen.r.Float32() < .5)
			}, // DNJ TODO - Generics and/or pass context for error?
			protoreflect.StringKind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(randString(gen.r, 1, 5))
			},
			protoreflect.Int64Kind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(gen.r.Int63n(20) - 10)
			},
			protoreflect.Int32Kind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(gen.r.Int31n(20) - 10)
			},
			protoreflect.Uint32Kind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(gen.r.Uint32())
			},
			protoreflect.FloatKind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(gen.r.Float32()*2 - 1)
			},
			protoreflect.EnumKind: func(gen generatorSource) protoreflect.Value {
				enum := gen.field.Enum().Values()
				selected := gen.r.Intn(enum.Len())
				return protoreflect.ValueOf(enum.Get(selected).Number())
			},
			protoreflect.MessageKind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(inputGenerator(ctx, gen.r, gen.field.Message(), potentialInputs, gen)) // DNJ TODO - this depth lineage is too confusing
			},
		},
		NameGenerators: map[string]Generator{
			"pps_v2.Input.join":                           subinput,
			"pps_v2.Input.cross":                          subinput,
			"pps_v2.Input.union":                          subinput,
			"pps_v2.Input.group":                          subinput,
			"pps_v2.Input.cron":                           noValue, // DNJ TODO - handle multiple input types set - store info in GenSource
			"pps_v2.Input.pfs":                            pfsInput,
			"google.protobuf.Int64Value.value":            randInt64Positive, // DNJ TODO - read ancestry to decide?
			"google.protobuf.Duration.seconds":            randInt64Positive,
			"google.protobuf.Duration.nanos":              randNanos,
			"google.protobuf.Timestamp.seconds":           randInt64Positive,
			"google.protobuf.Timestamp.nanos":             randNanos,
			"pps_v2.CreatePipelineRequest.tf_job":         noValue,
			"pps_v2.CreatePipelineRequest.reprocess_spec": randReproccessSpec,
			"pps_v2.DatumSetSpec.size_bytes":              randInt64Positive,
			"pps_v2.DatumSetSpec.number":                  randInt64Positive,
			"pps_v2.Toleration.operator":                  randTolerationOperator,
			"pps_v2.CreatePipelineRequest.pod_spec":       randJson,
			"pps_v2.CreatePipelineRequest.pod_patch":      randJson,
			"pfs_v2.Repo.type":                            constantVal("user"),
			"pps_v2.PFSInput.repo_type":                   constantVal("user"),
			"pps_v2.PFSInput.branch":                      constantVal("master"),
			"pps_v2.PFSInput.trigger":                     noValue,
			"pps_v2.PFSInput.commit":                      noValue,
			"pps_v2.PFSInput.s3":                          noValue,
			"pps_v2.PFSInput.repo":                        existingString("pfs_v2.Repo.name"),
			"pps_v2.PFSInput.project":                     existingString("pfs_v2.Project.name"),
			"pps_v2.Service.type":                         constantVal("NodePort"),
			"pps_v2.Transform.secrets":                    noValue,
			"pps_v2.CreatePipelineRequest.dry_run":        noValue,
			"pps_v2.CreatePipelineRequest.determined":     noValue,
			"pfs_v2.SQLDatabaseEgress.secret":             noValue,
			"pps_v2.CreatePipelineRequest.spout":          noValue,
			"pps_v2.CreatePipelineRequest.s3_out":         noValue,
			"pps_v2.CreatePipelineRequest.egress":         noValue,
			"pps_v2.CreatePipelineRequest.spec_commit":    noValue,
			"pps_v2.Pipeline.name":                        randName,
			"pps_v2.ParallelismSpec.constant":             constantVal(uint64(1)),
		},
	}
	for i := 0; i < msgDesc.Fields().Len(); i++ { // DNJ TODO fuzz each type and recurse
		field := msgDesc.Fields().Get(i)
		var fieldVal protoreflect.Value
		successInputsMu.Lock()
		if successVals, ok := potentialInputs[string(field.FullName())]; ok && // DNJ TODO figure out input naming
			r.Float32() < useSucccessValWeight {
			fieldVal = successVals[r.Intn(len(successVals))]
			successInputsMu.Unlock()
		} else {
			successInputsMu.Unlock()
			genSource := generatorSource{
				ctx:           ctx,
				r:             r,
				field:         field,
				index:         defaultGenerators,
				hasInputLevel: prevSource.hasInputLevel,
			}
			genSource.ancestry = append(prevSource.ancestry, field)
			if field.IsList() { // fall-back to normal assignment if 1 or fewer selected
				count := r.Intn(maxRepeatedCount) // DNJ TODO how to handle input level here? move list handling to a type function to manually handle??
				vals := msg.NewField(field).List()
				if !strings.Contains(string(field.FullName()), "pps_v2.Input") || *genSource.hasInputLevel < len(genSource.ancestry) { // DNJ TODO how to fix 1-off handling
					for i := 0; i < count; i++ {
						v := getValueForField(ctx, genSource)
						if v.IsValid() {
							vals.Append(v)
						} else {
							log.Info(ctx, "Value in a list was not valid! ignoring.",
								zap.Any("field", field.FullName()),
								zap.Any("value", v.Interface()))
						}
					}
				}
				fieldVal = protoreflect.ValueOf(vals)
			} else if field.IsMap() {
				log.Info(ctx, "Map message type not yet implemented in value generation",
					zap.Any("field", field.FullName()))
			} else {
				fieldVal = getValueForField(ctx, genSource)
			}
		}
		if string(field.Name()) == "operator" {
			log.Info(ctx, "DNJ TODO Field for input what?!",
				zap.Any("field", field.FullName()),
				zap.Any("fieldVal", fieldVal.Interface()))
		}
		if fieldVal.IsValid() {
			msg.Set(field, fieldVal)
		}
	}
	return msg
}

func getValueForField(ctx context.Context, genSource generatorSource) protoreflect.Value {
	if nameGenFunc, ok := genSource.index.NameGenerators[string(genSource.field.FullName())]; ok {
		return nameGenFunc(genSource)
	} else if kindGenFunc, ok := genSource.index.KindGenerators[genSource.field.Kind()]; ok {
		// log.Info(ctx, "DNJ TODO - field kind log",
		// 	zap.Any("field", genSource.field.FullName()),
		// 	zap.Any("field", genSource.field.Kind()),
		// )
		return kindGenFunc(genSource)
	} else {
		log.Info(ctx, "Input generator not provided for requested field.",
			zap.Any("field", genSource.field.FullName()),
			zap.Any("data type", genSource.field.Kind()))
		return protoreflect.ValueOf(nil)
	}
}

func storeSuccessVals(ctx context.Context, msg *dynamicpb.Message, out map[string][]protoreflect.Value) {
	// DNJ TODO expander
	msg.Range(func(field protoreflect.FieldDescriptor, fieldVal protoreflect.Value) bool {
		fieldName := string(field.FullName())
		successInputsMu.Lock()
		if vals, ok := out[fieldName]; ok {
			if !slices.ContainsFunc(vals, func(v protoreflect.Value) bool { return v.Interface() == fieldVal.Interface() }) { // avoid long lists of duplicates
				out[fieldName] = append(vals, fieldVal) // DNJ TODO - test
			}
		} else {
			out[fieldName] = []protoreflect.Value{fieldVal} // DNJ TODO repeated fields
		}
		successInputsMu.Unlock()
		if field.Kind() == protoreflect.MessageKind { // handle sub messages as well as saving the whole message
			subMsg, ok := fieldVal.Interface().(*dynamicpb.Message)
			if field.IsList() {
				list := fieldVal.List()
				for i := 0; i < list.Len(); i++ {
					listMsg, ok := list.Get(i).Message().(*dynamicpb.Message)
					if !ok {
						log.Info(ctx, "Storing a list of message values, but a the message was not a valid dynamicpb.Message", // DNJ TODO -error?
							zap.Any("field", field.FullName()),
							zap.Any("msg", list.Get(i).String()))
					} else {
						storeSuccessVals(ctx, listMsg, out)
					}
				}
			} else if field.IsMap() {
				log.Info(ctx, "Saving the Map message type not yet implemented",
					zap.Any("field", field.FullName()))
			} else if ok { // just a single message to recurse through
				storeSuccessVals(ctx, subMsg, out)
			}
		}
		return true // always continue since there's no error case
	})

}

func fuzzGrpc(ctx context.Context, c *client.APIClient, seed int64) { // DNJ TODO - can and/or should this be public?
	r := rand.New(rand.NewSource(seed)) // DNJ TODO - accept intrface that Rand stisifies to allow custom value sourcing? move successvals logic there! needs Float32 and Intn right now
	rangeRPCsList(protosUnderTest, func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) {
		input := inputGenerator(ctx, r, md.Input(), successInputs, generatorSource{hasInputLevel: new(int)})
		reply := dynamicpb.NewMessage(md.Output())
		fullName := "/" + path.Join(string(sd.FullName()), string(md.Name()))      // DNJ TODO move to rpc fuzz test and make function for this.
		if err := c.ClientConn().Invoke(ctx, fullName, input, reply); err != nil { // DNJ TODO - do in goroutine so we don't wait for response?
			log.Info(ctx, "Invoke grpc method err",
				zap.String("method name", fullName),
				zap.String("input", input.String()),
				zap.Error(err),
			)
		} else {
			log.Info(ctx, "successful rpc call, saving fields for re-use",
				zap.String("method name", fullName),
				zap.String("input", input.String()),
				zap.String("reply", reply.String()),
			)
			storeSuccessVals(ctx, input, successInputs)
			storeSuccessVals(ctx, reply, successInputs)
		}
		// t.Logf("Reply: %#v", reply)
	})
}

// DNJ TODO context and logger?
func TestGrpcWorkflow(t *testing.T) { // DNJ TODO - better name
	if !*runDefault {
		t.Skip()
	}
	// We need a file logger since Go runs Fuzz in seperate procs
	flushLog := log.InitBatchLogger("/tmp/FuzzGrpc.log")
	ctx := pctx.Background("FuzzGrpcWorkflow")
	c := getSharedCluster(t)
	// for i := 0; i < 100; i++ {
	// 	fuzzGrpc(f, ctx, c, time.Now().Unix())
	// 	time.Sleep(time.Millisecond * 1000)
	// }
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	// f.Fuzz(
	for i := 0; i < 10000; i++ {
		eg.Go(func() error {
			fuzzGrpc(ctx, c, rand.Int63())
			return nil
			// log.Info(ctx, "Stored input", zap.String("success inputs", fmt.Sprintf("%#v", successInputs)))
		})
	}

	flushLog(nil) // don't exit with error, log file will be non-empty and saved if there's anything to say
	//DNJ TODO validation in Fuzz()?
	if err := c.Fsck(true, func(*pfs.FsckResponse) error { return nil }); err == nil {
		require.NoError(t, err, "fsck should not error after fuzzing")
	}
}
