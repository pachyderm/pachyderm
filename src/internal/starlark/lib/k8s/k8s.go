// Package k8s is a Kubernetes API binding for Starlark.
package k8s

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
)

// resource is a starlark.Value wrapper around unstructured.Unstructured.  It is used over
// ourstar.Value(Unstructured) to preserve MarshalJSON when calling json.encode(obj) from Starlark.
type resource struct {
	*unstructured.Unstructured
}

var _ json.Marshaler = (*resource)(nil)
var _ ourstar.ToGoer = (*resource)(nil)
var _ starlark.Value = (*resource)(nil)
var _ starlark.Mapping = (*resource)(nil)
var _ starlark.Unpacker = (*resource)(nil)

func (r *resource) MarshalJSON() ([]byte, error) {
	if r.Unstructured == nil {
		return nil, nil
	}
	js, err := r.Unstructured.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "marshal underlying object")
	}
	return js, nil
}
func (r *resource) ToGo() any {
	return r.Unstructured
}
func (r *resource) String() string {
	result, err := r.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("%#v", r.Unstructured)
	}
	return string(result)
}
func (r *resource) Type() string {
	return "<" + r.Unstructured.GetObjectKind().GroupVersionKind().String() + ">"
}
func (r *resource) Freeze() {}
func (r *resource) Truth() starlark.Bool {
	return r.Unstructured != nil && len(r.Unstructured.Object) > 0
}
func (r *resource) Hash() (uint32, error) { return 0, errors.New("unhashable") }
func (r *resource) Get(x starlark.Value) (starlark.Value, bool, error) {
	k, ok := starlark.AsString(x)
	if !ok {
		return nil, false, errors.Errorf("value %v cannot be used as a string key", x)
	}
	if r.Unstructured == nil {
		return starlark.None, false, nil
	}
	v, ok := r.Unstructured.Object[k]
	if !ok {
		return starlark.None, false, nil
	}
	return ourstar.Value(v), true, nil
}
func (r *resource) Unpack(v starlark.Value) error {
	switch x := v.(type) {
	case starlark.String:
		us := unstructured.Unstructured{}
		if err := us.UnmarshalJSON([]byte(x)); err != nil {
			return errors.Wrap(err, "unmarshal")
		}
		return nil
	case starlark.Bytes:
		us := unstructured.Unstructured{}
		if err := us.UnmarshalJSON([]byte(x)); err != nil {
			return errors.Wrap(err, "unmarshal")
		}
		return nil
	case *starlark.Dict:
		obj := make(map[string]any)
		for _, item := range x.Items() {
			if len(item) != 2 {
				return errors.New("invalid tuple in dict item")
			}
			k, v := item[0], item[1]
			key, ok := starlark.AsString(k)
			if !ok {
				return errors.Errorf("convert %v to string", key)
			}
			obj[key] = ourstar.FromStarlark(v)
		}
		r.Unstructured = &unstructured.Unstructured{
			Object: obj,
		}
		return nil
	}
	return errors.Errorf("invalid type %#v for resource", v)
}

// resource is a starlark.Value wrapper around unstructured.UnstructuredList, where the unstructured
// list is actually a single object.  It is used over
// starlark.List(ourstar.Value(UnstructuredList.Item[0]), ...) to preserve MarshalJSON when calling
// json.encode(obj) from Starlark.
type resourceList struct {
	*unstructured.UnstructuredList
}

var _ json.Marshaler = (*resourceList)(nil)
var _ ourstar.ToGoer = (*resourceList)(nil)
var _ starlark.Value = (*resourceList)(nil)
var _ starlark.Mapping = (*resourceList)(nil)
var _ starlark.Iterable = (*resourceList)(nil)

func (r *resourceList) MarshalJSON() ([]byte, error) {
	if r == nil || r.UnstructuredList == nil {
		return nil, nil
	}
	js, err := r.UnstructuredList.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "marshal underlying list")
	}
	return js, nil
}
func (r *resourceList) ToGo() any {
	return r.UnstructuredList
}
func (r *resourceList) String() string {
	result, err := r.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("%#v", r.UnstructuredList)
	}
	return string(result)
}
func (r *resourceList) Type() string {
	return "<" + r.UnstructuredList.GetObjectKind().GroupVersionKind().String() + ">"
}
func (r *resourceList) Freeze() {}
func (r *resourceList) Truth() starlark.Bool {
	return r.UnstructuredList != nil && len(r.UnstructuredList.Items) > 0
}
func (r *resourceList) Hash() (uint32, error) { return 0, errors.New("unhashable") }

// Get treats the list of items as a map from name to value.
func (r *resourceList) Get(x starlark.Value) (starlark.Value, bool, error) {
	if r.UnstructuredList == nil {
		return starlark.None, false, nil
	}
	for _, v := range r.UnstructuredList.Items {
		rawmd, ok := v.Object["metadata"]
		if !ok {
			continue
		}
		md, ok := rawmd.(map[string]any)
		if !ok {
			continue
		}
		n, ok := md["name"]
		if !ok {
			continue
		}
		found, err := starlark.Compare(syntax.EQL, x, ourstar.Value(n))
		if err != nil {
			return starlark.None, false, errors.Wrapf(err, "compare %v to %v", x, n)
		}
		if !found {
			continue
		}
		return &resource{
			Unstructured: &unstructured.Unstructured{
				Object: v.Object,
			},
		}, true, nil
	}
	return starlark.None, false, nil
}
func (r *resourceList) Iterate() starlark.Iterator {
	var values []starlark.Value
	if r.UnstructuredList != nil {
		for _, x := range r.Items {
			x := x
			values = append(values, &resource{Unstructured: &x})
		}
	}
	return starlark.NewList(values).Iterate()
}

// wrappedError is an error pretending to be a k8s object.  Since scripts cannot handle errors, this
// lets them treat an error like it's an object.  The debug dump machinery, for example, knows how
// to write an error object, so if you're just doing something like:
//
//	pods = k8s.pods.list()
//	for pod in pods:
//	    dump(pod["metadata"]["name"]+".json", json.encode(pod))
//
// and list returns an error, the script will still complete, but you'll get a file named "error"
// instead of a JSON file for each pod.
type wrappedError struct {
	err error
}

var _ error = (*wrappedError)(nil)
var _ json.Marshaler = (*wrappedError)(nil)
var _ starlark.Value = (*wrappedError)(nil)
var _ starlark.Mapping = (*wrappedError)(nil)
var _ starlark.IterableMapping = (*wrappedError)(nil)
var _ starlark.Iterable = (*wrappedError)(nil)

func (e *wrappedError) Error() string         { return e.err.Error() }
func (e *wrappedError) String() string        { return fmt.Sprintf("Error: %v", e.err) }
func (e *wrappedError) Type() string          { return "error" }
func (e *wrappedError) Freeze()               {}
func (e *wrappedError) Truth() starlark.Bool  { return false }
func (e *wrappedError) Hash() (uint32, error) { return uint32(xxh3.HashString(e.err.Error())), nil }
func (e *wrappedError) Get(x starlark.Value) (starlark.Value, bool, error) {
	return e, true, nil
}
func (e *wrappedError) Iterate() starlark.Iterator { return &emptyIterator{} }
func (e *wrappedError) Items() []starlark.Tuple    { return nil }
func (e *wrappedError) MarshalJSON() ([]byte, error) {
	js, err := json.Marshal(map[string]string{"error": fmt.Sprintf("%+v", e.err)})
	if err != nil {
		return nil, errors.Wrap(err, "marshal error message")
	}
	return js, nil
}

type emptyIterator struct{}

var _ starlark.Iterator = emptyIterator{}

func (i emptyIterator) Next(p *starlark.Value) bool { return false }
func (i emptyIterator) Done()                       {}

// selector is a labels.Selector, from a Starlark string or dict.
type selector struct{ labels.Selector }

func (s *selector) String() string {
	if s == nil || s.Selector == nil {
		return ""
	}
	return s.Selector.String()
}
func (s *selector) Unpack(v starlark.Value) error {
	switch x := v.(type) {
	case starlark.String:
		l, err := labels.Parse(string(x))
		if err != nil {
			return errors.Wrap(err, "parse selector as string")
		}
		s.Selector = l
		return nil
	case *starlark.Dict:
		set := make(labels.Set)
		for _, item := range x.Items() {
			if len(item) != 2 {
				return errors.New("invalid item tuple in selector dict")
			}
			k, ok := starlark.AsString(item[0])
			if !ok {
				return errors.Errorf("invalid selector key %v", item[0])
			}
			v, ok := starlark.AsString(item[1])
			if !ok {
				return errors.Errorf("invalid selector value %v", item[1])
			}
			set[k] = v
		}
		s.Selector = set.AsSelector()
		return nil
	}
	return errors.Errorf("invalid selector %#v", v)
}

// resourceClient is a k8s client for a particular resource.  Each client has attributes like
// client.get(name) and client.list().
type resourceClient struct {
	resource  metav1.APIResource
	namespace string
	client    dynamic.Interface
}

var _ starlark.Value = (*resourceClient)(nil)
var _ starlark.HasAttrs = (*resourceClient)(nil)

func (c *resourceClient) verbs() []string {
	var verbs []string
	for _, v := range c.resource.Verbs {
		if v == "list" || v == "get" {
			verbs = append(verbs, v)
		}
	}
	sort.Strings(verbs)
	return verbs
}
func (c *resourceClient) String() string {
	b := new(strings.Builder)
	b.WriteRune('<')
	fmt.Fprintf(b, "%v_client", c.resource.Name)
	if c.resource.Namespaced {
		fmt.Fprintf(b, " namespace=%q", c.namespace)
	}
	if len(c.resource.Verbs) > 0 {
		fmt.Fprintf(b, " verbs=%v", c.verbs())
	}
	b.WriteRune('>')
	return b.String()
}
func (c *resourceClient) Type() string         { return fmt.Sprintf("%v_client", c.resource.Name) }
func (c *resourceClient) Freeze()              {}
func (c *resourceClient) Truth() starlark.Bool { return true }
func (c *resourceClient) Hash() (uint32, error) {
	return uint32(xxh3.HashString(c.resource.Name)), nil
}

func (c *resourceClient) AttrNames() []string {
	return c.verbs()
}

func (c *resourceClient) Attr(attr string) (starlark.Value, error) {
	if attr == "get" && c.can("get") {
		return c.makeGet(), nil
	}
	if attr == "list" && c.can("list") {
		return c.makeList(), nil
	}
	return nil, nil
}

func (c *resourceClient) can(x string) bool {
	for _, v := range c.resource.Verbs {
		if v == x {
			return true
		}
	}
	return false
}

func (c *resourceClient) getClient() dynamic.ResourceInterface {
	gvr := schema.GroupVersionResource{
		Group:    c.resource.Group,
		Version:  c.resource.Version,
		Resource: c.resource.Name,
	}
	cl := c.client.Resource(gvr)
	if c.resource.Namespaced {
		return cl.Namespace(c.namespace)
	}
	return cl
}

func (c *resourceClient) makeGet() *starlark.Builtin {
	return starlark.NewBuiltin(fmt.Sprintf("%v.%v", c.resource.Name, "get"), func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var name, resourceVersion string
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "resourceVersion?", &resourceVersion); err != nil {
			return nil, errors.Wrap(err, "unpack args")
		}
		ctx := ourstar.GetContext(t)
		got, err := c.getClient().Get(ctx, name, metav1.GetOptions{ResourceVersion: resourceVersion})
		if err != nil {
			return &wrappedError{err: err}, nil
		}
		return &resource{Unstructured: got}, nil
	})
}

func (c *resourceClient) makeList() *starlark.Builtin {
	return starlark.NewBuiltin(fmt.Sprintf("%v.%v", c.resource.Name, "list"), func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var labels, fields selector
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "labels?", &labels, "fields?", &fields); err != nil {
			return nil, errors.Wrap(err, "unpack args")
		}
		ctx := ourstar.GetContext(t)
		got, err := c.getClient().List(ctx, metav1.ListOptions{
			LabelSelector: labels.String(),
			FieldSelector: fields.String(),
		})
		if err != nil {
			return starlark.NewList([]starlark.Value{&wrappedError{err: err}}), nil
		}
		return &resourceList{UnstructuredList: got}, nil
	})
}

func NewClientset(namespace string, sc kubernetes.Interface, dc dynamic.Interface) (*starlarkstruct.Module, error) {
	// Discovery annoyingly doesn't take a context, so we cross our fingers that this doesn't
	// run forever.
	_, rls, err := sc.Discovery().ServerGroupsAndResources()
	if err != nil {
		return nil, errors.Wrap(err, "discover server resources")
	}
	result := &starlarkstruct.Module{
		Name: "k8s",
		Members: starlark.StringDict{
			// resource(object) returns an object from a string or dict for testing.
			"resource": starlark.NewBuiltin("resource", func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				var object resource
				if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "object", &object); err != nil {
					return nil, errors.Wrap(err, "unpack args")
				}
				return &object, nil
			}),

			// logs(name, previous=None, container=None) returns 10MB of pod logs.
			"logs": starlark.NewBuiltin("logs", func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				var name, container string
				var previous bool
				if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "container", &container, "previous?", &previous); err != nil {
					return nil, errors.Wrap(err, "unpack args")
				}
				opts := &v1.PodLogOptions{
					Container:  container,
					LimitBytes: ptr.To[int64](10 * units.MB),
					Previous:   previous,
				}
				ctx := ourstar.GetContext(t)
				bytes, err := sc.CoreV1().Pods(namespace).GetLogs(name, opts).DoRaw(ctx)
				if err != nil {
					return &wrappedError{err: err}, nil
				}
				return starlark.Bytes(bytes), nil
			}),
		},
	}
	for _, rl := range rls {
		gv, err := schema.ParseGroupVersion(rl.GroupVersion)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse GroupVersion received from the server")
		}
		for _, r := range rl.APIResources {
			// NOTE(jrockway): metrics.k8s.io seems broken on my machine; that's what
			// this tries to skip.
			if r.SingularName == "" {
				continue
			}

			var name string
			if parts := strings.SplitN(r.Name, "/", 2); len(parts) == 2 {
				name = parts[0]
			} else {
				name = r.Name
			}
			val, ok := result.Members[name]
			if !ok {
				val = &resourceClient{
					client:    dc,
					namespace: namespace,
				}
				result.Members[name] = val
			}
			client := val.(*resourceClient)
			if name == r.Name {
				r.Group = gv.Group
				r.Version = gv.Version
				client.resource = r
			}
			// Otherwise: the resource is a subresource (like pods/log), which discovery
			// returns wrong information for (it says that everything has get/list,
			// which is not true; the scale resources are update/patch, the log resource
			// is get, etc.), and the dynamic client is completely broken for (it can't
			// take parameters, and attempts to unmarshal everything as JSON, which
			// doesn't work for logs).
		}
	}
	return result, nil
}
