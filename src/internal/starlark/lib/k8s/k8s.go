// Package k8s is a Kubernetes API binding for Starlark.
package k8s

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type resourceClient struct {
	resource    metav1.APIResource
	subresource string
	extra       []metav1.APIResource
	namespace   string
	client      dynamic.Interface
}

var _ starlark.Value = (*resourceClient)(nil)
var _ starlark.HasAttrs = (*resourceClient)(nil)

func (c *resourceClient) shorten(r metav1.APIResource) string {
	return strings.TrimPrefix(r.Name, c.resource.Name+"/")
}
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
	if c.subresource == "" {
		fmt.Fprintf(b, "<%v_client", c.resource.Name)
	} else {
		fmt.Fprintf(b, "<%v.%v_client", c.resource.Name, c.subresource)
	}
	if c.resource.Namespaced {
		fmt.Fprintf(b, " namespace=%q", c.namespace)
	}
	if len(c.extra) > 0 {
		var extras []string
		for _, e := range c.extra {
			extras = append(extras, c.shorten(e))
		}
		sort.Strings(extras)
		fmt.Fprintf(b, " subresources=%v", extras)
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
	result := c.verbs()
	for _, e := range c.extra {
		result = append(result, c.shorten(e))
	}
	return result
}

func (c *resourceClient) Attr(attr string) (starlark.Value, error) {
	for _, e := range c.extra {
		if c.shorten(e) == attr {
			return &resourceClient{
				client:      c.client,
				namespace:   c.namespace,
				subresource: c.shorten(e),
				resource:    c.resource,
			}, nil
		}
	}
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
		var sr []string
		if c.subresource != "" {
			sr = append(sr, c.subresource)
		}
		got, err := c.getClient().Get(ctx, name, metav1.GetOptions{ResourceVersion: resourceVersion}, sr...)
		if err != nil {
			return &wrappedError{err: err}, nil
		}
		return ourstar.Value(got.Object), nil
	})
}

func (c *resourceClient) makeList() *starlark.Builtin {
	return nil
}

type wrappedError struct {
	err error
}

var _ error = (*wrappedError)(nil)
var _ starlark.Value = (*wrappedError)(nil)

func (e *wrappedError) String() string        { return fmt.Sprintf("Error: %v", e.err) }
func (e *wrappedError) Type() string          { return "error" }
func (e *wrappedError) Freeze()               {}
func (e *wrappedError) Truth() starlark.Bool  { return true }
func (e *wrappedError) Hash() (uint32, error) { return uint32(xxh3.HashString(e.err.Error())), nil }
func (e *wrappedError) Error() string         { return e.err.Error() }

var camelCase = regexp.MustCompile(`[^A-Z][A-Z]`)

func set(dest map[string]map[string]map[string]*starlarkstruct.Module, gvk schema.GroupVersionKind, m *starlarkstruct.Module) {
	group, version, kind := gvk.Group, gvk.Version, gvk.Kind
	group = strings.ReplaceAll(group, ".", "_")
	version = strings.ReplaceAll(version, ".", "_")
	kind = strings.ReplaceAll(kind, "API", "api")
	kind = camelCase.ReplaceAllStringFunc(kind, func(s string) string {
		if len(s) == 2 {
			return strings.ToLower(string([]byte{s[0], '_', s[1]}))
		}
		return strings.ToLower(s)
	})
	kind = strings.ToLower(kind)
	if gvk.Group == "" { // Core objects.
		group, version, kind = version, kind, ""
	}
	if dest[group] == nil {
		dest[group] = make(map[string]map[string]*starlarkstruct.Module)
	}
	if dest[group][version] == nil {
		dest[group][version] = make(map[string]*starlarkstruct.Module)
	}
	dest[group][version][kind] = m
}

func NewClientset(namespace string, sc kubernetes.Interface, dc dynamic.Interface) (*starlarkstruct.Module, error) {
	_, rls, err := sc.Discovery().ServerGroupsAndResources()
	if err != nil {
		return nil, errors.Wrap(err, "discover server resources")
	}
	result := &starlarkstruct.Module{
		Name:    "k8s",
		Members: starlark.StringDict{},
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
			} else {
				client.extra = append(client.extra, r)
			}
		}
	}
	return result, nil
}
