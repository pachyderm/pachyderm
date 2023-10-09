package starlark

import (
	"fmt"
	"sort"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/spf13/pflag"
	"golang.org/x/exp/maps"
)

var (
	// Personalities are modes of running Starpach scripts, like "debug", "build", "test", etc.
	// Scripts of a certain personality may have predefined bindings for that personality; debug
	// has "dump()" predefined, for example.
	personalities = map[string]Options{}
)

// personality is a type that holds a personality set from the command line.
type personality struct {
	Name    string
	Options Options
}

var _ pflag.Value = (*personality)(nil)

func (f *personality) Set(x string) error {
	p, ok := personalities[x]
	if !ok {
		ps := maps.Keys(personalities)
		sort.Strings(ps)
		return errors.Errorf("no personality %v; try one of %v", x, ps)
	}
	f.Name = x
	f.Options = p
	return nil
}

func (f *personality) String() string {
	return f.Name
}

func (f *personality) Type() string {
	return "personality"
}

// Personality is the personality selected from the command line.
var Personality = personality{
	Name:    "",
	Options: Options{},
}

// RegisterPersonality registers a personality; intended to be called from the init() section of
// modules that provide starlark bindings.
func RegisterPersonality(name string, opts Options) {
	if _, ok := personalities[name]; ok {
		panic(fmt.Sprintf("personality %q already registered", name))
	}
	personalities[name] = opts
}

// UsePersonalityFlag registers DefaultPersonality with the provided flagset.
func UsePersonalityFlag(flagset *pflag.FlagSet) {
	ps := maps.Keys(personalities)
	sort.Strings(ps)
	flagset.VarP(&Personality, "personality", "p",
		fmt.Sprintf("Start the interpreter with the personality of a certain type of script; one of %v", ps))
}
