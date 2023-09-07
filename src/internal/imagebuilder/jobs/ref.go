package jobs

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"go.starlark.net/starlark"
)

// Reference is something that can match a build input or artifact.
type Reference interface{ Match(any) bool }
type WithName interface{ GetName() string }
type WithPlatform interface{ GetPlatform() Platform }
type ConstrainableToPlatform interface{ ConstrainToPlatform(Platform) Reference }

// ReferenceList is a list of references that behaves nicely in starlark.
type ReferenceList []Reference

var _ starlark.Value = (ReferenceList)(nil)
var _ starlark.Mapping = (ReferenceList)(nil)
var _ starlark.Indexable = (ReferenceList)(nil)
var _ starlark.Sliceable = (ReferenceList)(nil)
var _ starlark.Iterable = (ReferenceList)(nil)

func (ReferenceList) Freeze()               {} // Always frozen.
func (ReferenceList) Hash() (uint32, error) { return 0, errors.New("ReferenceList is unhashable") }
func (r ReferenceList) Truth() starlark.Bool {
	if len(r) > 0 {
		return starlark.True
	}
	return starlark.False
}
func (r ReferenceList) String() string {
	b := new(strings.Builder)
	b.WriteRune('[')
	for i, ref := range r {
		fmt.Fprintf(b, "%q", ref)
		if i < len(r)-1 {
			b.WriteString(", ") // repr has commas.
		}
	}
	b.WriteRune(']')
	return b.String()
}
func (r ReferenceList) Type() string { return "referencelist" }
func (r ReferenceList) Get(v starlark.Value) (starlark.Value, bool, error) {
	if x, err := starlark.NumberToInt(v); err == nil {
		if i, ok := x.Int64(); ok {
			return r.Index(int(i)), true, nil
		}
	}
	if x, ok := starlark.AsString(v); ok {
		ref, err := ParseRef(x)
		if err != nil {
			return nil, false, errors.Wrap(err, "unparseable reference")
		}
		var result ReferenceList
		for _, target := range r {
			if ref.Match(target) {
				result = append(result, target)
			}
		}
		return result, true, nil
	}
	return nil, false, errors.Errorf("cannot map reference list with %v", v)
}
func (r ReferenceList) Len() int                   { return len(r) }
func (r ReferenceList) Index(i int) starlark.Value { return refWrapper{Reference: r[i]} }

// From Hacker's Delight, section 2.8.
func signum64(x int64) int { return int(uint64(x>>63) | uint64(-x)>>63) }
func signum(x int) int     { return signum64(int64(x)) }

func (r ReferenceList) Slice(start, end, step int) starlark.Value {
	// Starlark slices are [start:end:step], not [start:end:cap]!
	var result ReferenceList
	sign := signum(step)
	for i := start; signum(end-i) == sign; i += step {
		result = append(result, r[i])
	}
	return result
}
func (r ReferenceList) Iterate() starlark.Iterator {
	var vs []starlark.Value
	for _, ref := range r {
		vs = append(vs, refWrapper{Reference: ref})
	}
	return starlark.NewList(vs).Iterate()
}

// refWrapper wraps a Reference in a starlark.Value.
type refWrapper struct {
	Reference
}

var _ starlark.Value = (*refWrapper)(nil)

func (refWrapper) Freeze()                {} // Always frozen.
func (refWrapper) Hash() (uint32, error)  { return 0, errors.New("refWrapper is unhashable") }
func (r refWrapper) Truth() starlark.Bool { return r.Reference != nil }
func (r refWrapper) String() string       { return fmt.Sprint(r.Reference) }
func (r refWrapper) Type() string         { return "reference" }

// Name references something by name.
type Name string

var _ Reference = Name("")
var _ WithName = Name("")
var _ ConstrainableToPlatform = Name("")

// NameAndPlatform references something by name and platform.
type NameAndPlatform struct {
	Name     string
	Platform Platform
}

var _ Reference = (*NameAndPlatform)(nil)
var _ WithName = (*NameAndPlatform)(nil)
var _ WithPlatform = (*NameAndPlatform)(nil)

func (me Name) GetName() string { return string(me) }

func (me Name) Match(target any) bool {
	if x, ok := target.(WithName); ok {
		return me.GetName() == x.GetName()
	}
	return me == target
}

func (me Name) ConstrainToPlatform(p Platform) Reference {
	return NameAndPlatform{
		Name:     me.GetName(),
		Platform: p,
	}
}

func (me Name) String() string {
	return me.GetName()
}

func (me NameAndPlatform) GetName() string       { return me.Name }
func (me NameAndPlatform) GetPlatform() Platform { return me.Platform }
func (me NameAndPlatform) Match(target any) bool {
	if me == target {
		return true
	}
	if me.GetPlatform() == AllPlatforms {
		return true
	}
	if x, ok := target.(WithName); !ok || me.GetName() != x.GetName() {
		return false
	}
	if x, ok := target.(WithPlatform); !ok || (me.GetPlatform() != x.GetPlatform()) {
		return false
	}
	return true
}

func (me NameAndPlatform) String() string {
	return me.Name + "#" + string(me.Platform)
}

// ParseRef parse a string into a reference.  The string is split into 2 parts at #, the stuff
// before matches the name, and the stuff after matches the platform.  If either is empty, the
// matcher does not match the part that's empty.  If both are empty, the matcher always matches.
func ParseRef(x string) (Reference, error) {
	parts := strings.SplitN(x, "#", 2)
	var n, p string
	switch {
	case len(parts) == 2:
		n, p = parts[0], parts[1]
	case len(parts) == 1:
		n = parts[0]
	}
	switch {
	case n != "" && p != "":
		return NameAndPlatform{
			Name:     n,
			Platform: Platform(p),
		}, nil
	case n == "" && p != "":
		return Platform(p), nil
	case n != "" && p == "":
		return Name(n), nil
	}
	return nil, errors.New("empty matcher")
}
