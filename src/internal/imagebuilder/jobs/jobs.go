package jobs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type Artifact = any

// Job is work to do.
type Job interface {
	// ID allows jobs to be inserted into a set.
	ID() uint64
	// Inputs are references to the inputs that Run desires.
	Inputs() []Reference
	// Outputs reference the objects produced by Run.  Every actual output must be referenced by
	// exactly one Outputs reference.
	Outputs() []Reference
	// Run produces the outputs from the inputs.
	Run(context.Context, *JobContext, []Artifact) ([]Artifact, error)
}

// StarlarkCommand is a job that can be created in Starlark.
type StarlarkCommand interface {
	NewFromStarlark(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) ([]Job, error)
}

// JobContext is runtime information about the job system.
type JobContext struct {
	allowedPushPrefixes []string
	Cache               *Cache
	HTTPClient          *http.Client
}

func (jc *JobContext) PushAllowed(path string) bool {
	for _, p := range jc.allowedPushPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// PrintJob prints out a dump of a job.
func PrintJob(w io.Writer, job Job) {
	fmt.Fprintf(w, "job of type %T %v:\n", job, job)
	for i, in := range job.Inputs() {
		fmt.Fprintf(w, "    %d: %v\n", i, in)
	}
	fmt.Fprintf(w, " outputs ->\n")
	for i, out := range job.Outputs() {
		fmt.Fprintf(w, "    %d: %v\n", i, out)
	}
}

const (
	StarlarkRegistryKey = "jobsRegistry"
)

func MakeStarlarkCommand[T StarlarkCommand](name string) starlark.Value {
	var job T
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		reg := GetRegistry(thread)
		if reg.frozen {
			return nil, errors.New("registry is frozen")
		}
		js, err := job.NewFromStarlark(thread, fn, args, kwargs)
		if err != nil {
			return nil, err
		}
		for _, j := range js {
			log.Debug(ourstar.GetContext(thread), "registered a new job via starlark", zap.Any("job", j))
		}
		var outputs ReferenceList
		for _, j := range js {
			reg.Jobs = append(reg.Jobs, j)
			for _, o := range j.Outputs() {
				outputs = append(outputs, o)
			}
		}
		if len(outputs) == 0 {
			return starlark.None, nil
		}
		return outputs, nil
	})
}

// The Registry stores every job registered in Starlark code.  This way, workflow authors do not
// have to retain multiple return values; they only need to connect references together.  Then
// later, they can use the references they held onto to run `registry.resolve(ref)` or similar.
//
// Go code should not use the registry.
type Registry struct {
	frozen bool
	Jobs   []Job
}

// GlobalRegistry is a fake value so users can write "registry.<method>" in the debug shell.  The
// actual registry is retrieved from thread-local storage inside each method.
type GlobalRegistry struct{}

func buildRef(v starlark.Value) ([]Reference, error) {
	var result []Reference
	if x, ok := v.(refWrapper); ok {
		return []Reference{x.Reference}, nil
	}
	if x, ok := v.(ReferenceList); ok {
		result = append(result, x...)
		return result, nil
	}
	if _, ok := v.(starlark.String); !ok {
		if x, ok := v.(starlark.Indexable); ok {
			n := x.Len()
			for i := 0; i < n; i++ {
				r, err := buildRef(x.Index(i))
				if err != nil {
					return nil, errors.Wrapf(err, "*starlark.List[%d]", i)
				}
				result = append(result, r...)
			}
			return result, nil
		}
	}
	if x, ok := starlark.AsString(v); ok {
		r, err := ParseRef(x)
		if err != nil {
			return nil, errors.Wrap(err, "parse info ref")
		}
		return []Reference{r}, nil
	}
	return nil, errors.Errorf("cannot convert %v to a ref", v)
}

func refArgs(args starlark.Tuple, kwargs []starlark.Tuple) ([]Reference, error) {
	if len(kwargs) > 0 {
		return nil, errors.New("unexpected kwargs")
	}
	var refs []Reference
	for i, arg := range args {
		result, err := buildRef(arg)
		if err != nil {
			return nil, errors.Wrapf(err, "arg %d", i)
		}
		refs = append(refs, result...)
	}
	return refs, nil
}

var globalRegistryMethods = map[string]starlark.Value{
	"plan": starlark.NewBuiltin("plan", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		refs, err := refArgs(args, kwargs)
		if err != nil {
			return nil, err
		}
		ctx := ourstar.GetContext(thread)
		reg := GetRegistry(thread)

		plan, err := Plan(ctx, reg.Jobs, refs)
		if err != nil {
			return nil, err
		}
		var result []starlark.Value
		for _, paragraph := range plan {
			var lines []starlark.Value
			for _, line := range paragraph {
				lines = append(lines, starlark.String(line))
			}
			result = append(result, starlark.NewList(lines))
		}
		return starlark.NewList(result), nil
	}),
	"resolve": starlark.NewBuiltin("resolve", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		refs, err := refArgs(args, kwargs)
		if err != nil {
			return nil, err
		}
		ctx := ourstar.GetContext(thread)
		reg := GetRegistry(thread)

		result, err := Resolve(ctx, reg.Jobs, refs)
		if err != nil {
			return nil, err
		}
		return ourstar.ReflectList(result), nil
	}),
	"jobs": starlark.NewBuiltin("jobs", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if len(args) > 0 {
			return nil, errors.New("unexpected args")
		}
		if len(kwargs) > 0 {
			return nil, errors.New("unexpected kwargs")
		}
		return ourstar.ReflectList(GetRegistry(thread).Jobs), nil
	}),
	"outputs": starlark.NewBuiltin("outputs", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if len(args) > 0 {
			return nil, errors.New("unexpected args")
		}
		if len(kwargs) > 0 {
			return nil, errors.New("unexpected kwargs")
		}
		reg := GetRegistry(thread)
		result, err := Outputs(reg.Jobs)
		if err != nil {
			return nil, errors.Wrap(err, "generate outputs")
		}
		return ReferenceList(result), nil
	}),
}

var _ starlark.Value = (*GlobalRegistry)(nil)
var _ starlark.HasAttrs = (*GlobalRegistry)(nil)

func (r *GlobalRegistry) String() string        { return "<global registry>" }
func (r *GlobalRegistry) Type() string          { return "registry" }
func (r *GlobalRegistry) Truth() starlark.Bool  { return true }
func (r *GlobalRegistry) Hash() (uint32, error) { return 0, errors.New("registry is unhashable") }
func (r *GlobalRegistry) Freeze()               {}
func (r *GlobalRegistry) Attr(name string) (starlark.Value, error) {
	method, ok := globalRegistryMethods[name]
	if !ok {
		return nil, errors.Errorf("no attr %v", name)
	}
	return method, nil
}
func (r *GlobalRegistry) AttrNames() []string { return maps.Keys(globalRegistryMethods) }

func GetRegistry(thread *starlark.Thread) *Registry {
	return thread.Local(StarlarkRegistryKey).(*Registry) // let this panic
}
