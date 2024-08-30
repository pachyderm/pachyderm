// Package jsonnet is a Gazelle plugin that generates rules when it sees jsonnet files in the
// repository.
package jsonnet

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type Jsonnet struct {
	language.BaseLang // Stubs to implement everything we don't care about.
	Verbose           bool
}

func NewLanguage() language.Language {
	return new(Jsonnet)
}

const (
	lang             = "jsonnet"
	lintRule         = "jsonnet_lint_test"
	ruleImport       = "//private/rules/jsonnet-lint:jsonnet-lint.bzl"
	verboseDirective = "jsonnet_verbose" // "# gazelle:jsonnet_verbose [true|false]" in BUILD file to control logging.
)

var (
	kindInfo = map[string]rule.KindInfo{
		lintRule: {
			MatchAny:   false,
			MatchAttrs: []string{"src"},
			NonEmptyAttrs: map[string]bool{
				"src": true,
			},
			SubstituteAttrs: map[string]bool{},
			MergeableAttrs: map[string]bool{
				"deps": true,
			},
			ResolveAttrs: map[string]bool{},
		},
	}
)

func (Jsonnet) Name() string { return lang }

func (Jsonnet) Kinds() map[string]rule.KindInfo {
	return kindInfo
}

func (Jsonnet) KnownDirectives() []string { return []string{verboseDirective} }

func (j *Jsonnet) Configure(c *config.Config, rel string, f *rule.File) {
	if f == nil {
		return
	}
	for _, d := range f.Directives {
		switch d.Key {
		case verboseDirective:
			v, err := strconv.ParseBool(d.Value)
			if err != nil {
				log.Printf("error: unknown %v directive value in %v: should be true or false, not %v", verboseDirective, filepath.Join(rel, filepath.Base(f.Path)), d.Value)
				continue
			}
			j.Verbose = v
		}
	}
}

func (Jsonnet) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	switch r.Kind() {
	case lintRule:
		return []resolve.ImportSpec{
			{Lang: lang, Imp: ruleImport},
		}
	}
	return nil
}

func (Jsonnet) ApparentLoads(moduleToApparentName func(string) string) []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name:    ruleImport,
			Symbols: []string{"jsonnet_lint_test"},
		},
	}
}

func (j *Jsonnet) GenerateRules(args language.GenerateArgs) (result language.GenerateResult) {
	entries, err := os.ReadDir(args.Dir)
	if err != nil {
		log.Printf("jsonnet: error: ReadDir(%v): %v", args.Dir, err)
		return
	}
	var deps []string
	var rules []*rule.Rule
	seen := make(map[string]struct{})

	// Find jsonnet and libsonnet files.
	for _, e := range entries {
		// Directories can't be jsonnet files.
		if e.IsDir() {
			continue
		}

		// Path is a pretty filename for verbose logs.
		fullPath := filepath.Join(args.Dir, e.Name())
		path, err := filepath.Rel(args.Config.RepoRoot, fullPath)
		if err != nil {
			log.Printf("jsonnet: error: filepath.Rel(%v, %v): %v", args.Config.RepoRoot, fullPath, err)
			path = "<unknown>/" + e.Name()
		}

		// Generate a rule name; also collect a libsonnet file as a dependency.
		var name string
		ext := filepath.Ext(path)
		switch ext {
		case ".jsonnet":
			name = e.Name()[:len(e.Name())-len(ext)]
		case ".libsonnet":
			name = e.Name()[:len(e.Name())-len(ext)] + "_libsonnet"
			// This is really naive dependency resolution; every jsonnet file in a given
			// directory depends on all libsonnet files in that directory.
			deps = append(deps, e.Name())
		default:
			continue
		}
		if j.Verbose {
			log.Printf("jsonnet: emitting rules for %v", path)
		}

		// Create the rule without deps.  Deps will be added after this loop, so we know
		// we've seen them all.
		ruleName := fmt.Sprintf("%s_lint_test", strings.ReplaceAll(name, "-", "_"))
		r := rule.NewRule(lintRule, ruleName)
		r.SetPrivateAttr("filename", e.Name())
		seen[e.Name()] = struct{}{}
		r.SetAttr("size", "small")
		r.SetAttr("src", e.Name())
		rules = append(rules, r)
	}

	// Emit the rules, merging in deps if there are any.
	for _, r := range rules {
		var localDeps []string
		for _, d := range deps {
			if d != r.PrivateAttr("filename") {
				// A file shouldn't depend on itself.  It wouldn't be harmful, just
				// confusing and ugly.
				localDeps = append(localDeps, d)
			}
		}
		if len(localDeps) > 0 {
			r.SetAttr("deps", localDeps)
		}
		result.Gen = append(result.Gen, r)
		result.Imports = append(result.Imports, r.PrivateAttr("filename"))
	}

	// Do a final pass over existing rules, to see if there are any that need to be deleted
	// because the underlying source file was deleted.
	if args.File != nil {
		for _, r := range args.File.Rules {
			if r.Kind() != lintRule {
				continue
			}
			src := r.AttrString("src")
			if _, ok := seen[src]; !ok {
				r.DelAttr("src")
				result.Empty = append(result.Empty, r)
				if j.Verbose {
					log.Printf("jsonnet: deleting dead rule in %v: %v(%v)", args.Dir, r.Name(), r.Kind())
				}
			}
		}
	}
	return
}
