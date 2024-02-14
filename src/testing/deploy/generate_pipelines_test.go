//go:build k8s

package main

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
	useSucccessValWeight = .75 // Likelyhood of using a pre-existing, known valid, value.
	maxRepeatedCount     = 5   // The maximum number of random list values to generate.
	nilInputWeight       = .35 // The likelyhood of using nil as a value in an Input field
	maxInputDepth        = 15  // The maximum depth we should traverse the cyclical Input graph.
)

var (
	runAmount       = flag.Int("gen.Amount", 250, "use the default data generators")
	successInputs   = map[string][]protoreflect.Value{} // values of fields mapped to field names for reuse in subsequent requests
	successInputsMu = sync.Mutex{}
)

var protosUnderTest = map[protoreflect.FileDescriptor][]string{
	pfs.File_pfs_pfs_proto: {
		pfs.API_CreateRepo_FullMethodName, // It would be nice to include DeleteRepo but we need to weight it less so we create more than we delete.
		pfs.API_CreateProject_FullMethodName,
		pfs.API_ListProject_FullMethodName,
		pfs.API_ListRepo_FullMethodName,
	},
	pps.File_pps_pps_proto: {
		pps.API_CreatePipeline_FullMethodName,
		pps.API_ListPipeline_FullMethodName,
	},
}

type GeneratorIndex struct {
	KindGenerators map[protoreflect.Kind]Generator
	NameGenerators map[string]Generator
}

type generatorSource struct {
	ctx   context.Context
	r     *rand.Rand
	field protoreflect.FieldDescriptor
	// The deepest current traversal of the cyclic Input graph.
	// If we are above this level when setting input we should always generate nil since there can be only one input.
	// This doesn't work in some edge cases and is pretty confusing. It would be nice to have a better solution here.
	hasInputLevel *int
	ancestry      []protoreflect.FieldDescriptor
	index         GeneratorIndex
}

type Generator func(generatorSource) protoreflect.Value

func subinput(genSource generatorSource) protoreflect.Value {
	if len(genSource.ancestry) > genSource.r.Intn(maxInputDepth) {
		return protoreflect.ValueOf(nil) // have to terminate eventually in the input chain
	} else {
		*genSource.hasInputLevel = len(genSource.ancestry)
		return protoreflect.ValueOf(
			inputGenerator(
				genSource.ctx,
				genSource.r,
				genSource.field.Message(),
				successInputs,
				genSource,
			),
		)
	}
}

func pfsInput(genSource generatorSource) protoreflect.Value {
	ancestrylen := len(genSource.ancestry)
	if *genSource.hasInputLevel >= ancestrylen || nilInputWeight > genSource.r.Float32() {
		return protoreflect.ValueOf(nil)
	} else {
		*genSource.hasInputLevel = ancestrylen // Only one input per level allowed
		return protoreflect.ValueOf(
			inputGenerator(
				genSource.ctx,
				genSource.r,
				genSource.field.Message(),
				successInputs,
				genSource,
			),
		)
	}
}

func noValue(genSource generatorSource) protoreflect.Value {
	return protoreflect.ValueOf(nil)
}

func constantVal(value interface{}) Generator {
	return func(genSource generatorSource) protoreflect.Value { return protoreflect.ValueOf(value) }
}

func randInt64Positive(genSource generatorSource) protoreflect.Value {
	return protoreflect.ValueOf(genSource.r.Int63n(2000))
}

func randNanos(genSource generatorSource) protoreflect.Value {
	return protoreflect.ValueOf(genSource.r.Int31n(2000))
}

func randJson(genSource generatorSource) protoreflect.Value {
	return protoreflect.ValueOf("")
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

func rangeRPCsList(protos map[protoreflect.FileDescriptor][]string, f func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor)) {
	for fd, protoDescriptors := range protos {
		svcMethods := map[string][]string{}
		for _, name := range protoDescriptors {
			split := strings.Split(name, "/")
			svcName := split[1] // 0 is empty
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

func randString(r *rand.Rand, minlength int, maxLength int) string {
	var characters strings.Builder
	length := r.Intn(maxLength) + minlength
	for i := 0; i < length; i++ {
		characters.WriteByte(byte(32 + r.Intn(94))) // printable ascii
	}
	return characters.String()
}

func inputGenerator(ctx context.Context, r *rand.Rand, msgDesc protoreflect.MessageDescriptor, potentialInputs map[string][]protoreflect.Value, prevSource generatorSource) *dynamicpb.Message {
	msg := dynamicpb.NewMessage(msgDesc)
	defaultGenerators := GeneratorIndex{
		KindGenerators: map[protoreflect.Kind]Generator{
			protoreflect.BoolKind: func(gen generatorSource) protoreflect.Value {
				return protoreflect.ValueOf(gen.r.Float32() < .5)
			},
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
				return protoreflect.ValueOf(inputGenerator(ctx, gen.r, gen.field.Message(), potentialInputs, gen))
			},
		},
		NameGenerators: map[string]Generator{
			"pps_v2.Input.join":                           subinput,
			"pps_v2.Input.cross":                          subinput,
			"pps_v2.Input.union":                          subinput,
			"pps_v2.Input.group":                          subinput,
			"pps_v2.Input.cron":                           noValue,
			"pps_v2.Input.pfs":                            pfsInput,
			"google.protobuf.Int64Value.value":            randInt64Positive,
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
	for i := 0; i < msgDesc.Fields().Len(); i++ {
		field := msgDesc.Fields().Get(i)
		var fieldVal protoreflect.Value
		successInputsMu.Lock()
		if successVals, ok := potentialInputs[string(field.FullName())]; ok &&
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
				count := r.Intn(maxRepeatedCount)
				vals := msg.NewField(field).List()
				// we have one-off handling for input here. This should be re-worked to handle this more generically in the generators.
				if !strings.Contains(string(field.FullName()), "pps_v2.Input") || *genSource.hasInputLevel < len(genSource.ancestry) {
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
		return kindGenFunc(genSource)
	} else {
		log.Info(ctx, "Input generator not provided for requested field.",
			zap.Any("field", genSource.field.FullName()),
			zap.Any("data type", genSource.field.Kind()))
		return protoreflect.ValueOf(nil)
	}
}

// store successful values from requet inputs and outputs to be reused later.
func storeSuccessVals(ctx context.Context, msg *dynamicpb.Message, out map[string][]protoreflect.Value) {
	msg.Range(func(field protoreflect.FieldDescriptor, fieldVal protoreflect.Value) bool {
		fieldName := string(field.FullName())
		successInputsMu.Lock()
		if vals, ok := out[fieldName]; ok {
			if !slices.ContainsFunc(vals, func(v protoreflect.Value) bool { return v.Interface() == fieldVal.Interface() }) { // avoid long lists of duplicates
				out[fieldName] = append(vals, fieldVal)
			}
		} else {
			out[fieldName] = []protoreflect.Value{fieldVal}
		}
		successInputsMu.Unlock()
		if field.Kind() == protoreflect.MessageKind { // handle sub messages as well as saving the whole message
			subMsg, ok := fieldVal.Interface().(*dynamicpb.Message)
			if field.IsList() {
				list := fieldVal.List()
				for i := 0; i < list.Len(); i++ {
					listMsg, ok := list.Get(i).Message().(*dynamicpb.Message)
					if !ok {
						log.Info(ctx, "Storing a list of message values, but a the message was not a valid dynamicpb.Message",
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

// generate and send requests for each API in the proto list once.
func fuzzGrpc(ctx context.Context, c *client.APIClient, seed int64) {
	r := rand.New(rand.NewSource(seed))
	rangeRPCsList(protosUnderTest, func(fd protoreflect.FileDescriptor, sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) {
		input := inputGenerator(ctx, r, md.Input(), successInputs, generatorSource{hasInputLevel: new(int)})
		reply := dynamicpb.NewMessage(md.Output())
		fullName := "/" + path.Join(string(sd.FullName()), string(md.Name()))
		if err := c.ClientConn().Invoke(ctx, fullName, input, reply); err != nil {
			log.Info(ctx, "Invoke grpc method had an  err",
				zap.Error(err),
				zap.String("method name", fullName),
				zap.String("input", input.String()),
			)
		} else {
			log.Info(ctx, "successful rpc call, saving fields for re-use",
				zap.String("method name", fullName),
				zap.String("reply", reply.String()),
				zap.String("input", input.String()),
			)
			storeSuccessVals(ctx, input, successInputs)
			storeSuccessVals(ctx, reply, successInputs)
		}
	})
}

// Test generating random dags and verifying they get through the upgrade and pass fsck.
// We aren't using Fuzz* because that works better with stateless fuzzing. Here, the whole point
// is to add interesting state tot the DB.
func TestCreateDags(t *testing.T) {
	ra := *runAmount
	if ra <= 0 || testing.Short() {
		t.Skip("Skipping DAG generation test")
	}
	flushLog := log.InitBatchLogger("/tmp/FuzzGrpc.log")
	ctx := pctx.Background("FuzzGrpcWorkflow")

	deployOpts := &minikubetestenv.DeployOpts{
		Version:            "2.8.3",
		CleanupAfter:       false,
		UseLeftoverCluster: false,
		ValueOverrides: map[string]string{
			"prxoy.resources.requests.memory": "1Gi",
			"prxoy.resources.limits.memory":   "1Gi",
			"proxy.replicas":                  "5", // proxy overload_manager oom kills incoming requests to avoid taking down the cluster, which is smart, but not what we want here.
			"console.enabled":                 "true",
			"pachd.enterpriseLicenseKey":      os.Getenv("ENT_ACT_CODE"),
			"pachd.activateAuth":              "false",
		},
	}
	k := testutil.GetKubeClient(t)
	namespace, _ := minikubetestenv.ClaimCluster(t)
	c := minikubetestenv.InstallRelease(t, context.Background(), namespace, k, deployOpts)

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	for i := 0; i < ra; i++ {
		eg.Go(func() error {
			fuzzGrpc(ctx, c, rand.Int63())
			return nil
		})
	}
	if err := c.Fsck(true, func(*pfs.FsckResponse) error { return nil }); err == nil {
		require.NoError(t, err, "fsck should not error after fuzzing")
	}
	deployOpts.Version = "" // use the default (this commit in CI)
	minikubetestenv.UpgradeRelease(t, ctx, namespace, testutil.GetKubeClient(t), deployOpts)
	if err := c.Fsck(true, func(*pfs.FsckResponse) error { return nil }); err == nil {
		require.NoError(t, err, "fsck should not error after upgrade")
	}
	for i := 0; i < ra; i++ {
		eg.Go(func() error {
			fuzzGrpc(ctx, c, rand.Int63())
			return nil
		})
	}
	if err := c.Fsck(true, func(*pfs.FsckResponse) error { return nil }); err == nil {
		require.NoError(t, err, "fsck should not error after upgrade and fuzzing")
	}
	flushLog(nil) // don't exit with error, log file will be non-empty and saved if there's anything to say
}
