// Package jsonnet is a Gazelle plugin that generates rules when it sees jsonnet files in the
// repository.
package jsonnet

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

const (
	lang       = "jsonnet"
	lintRule   = "jsonnet_lint_test"
	ruleImport = "//private/rules/jsonnet-lint:jsonnet-lint.bzl"
)

func (_ Jsonnet) Name() string { return lang }

func (_ Jsonnet) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
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
}

func (_ Jsonnet) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	log.Printf("looking at rule %v (%v)", r.Name(), r.Kind())
	switch r.Kind() {
	case lintRule:
		return []resolve.ImportSpec{
			{Lang: lang, Imp: ruleImport},
		}
	}
	return nil
}

func (_ Jsonnet) ApparentLoads(moduleToApparentName func(string) string) []rule.LoadInfo {
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

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fullPath := filepath.Join(args.Dir, e.Name())
		path, err := filepath.Rel(args.Config.RepoRoot, fullPath)
		if err != nil {
			log.Printf("jsonnet: error: filepath.Rel(%v, %v): %v", args.Config.RepoRoot, fullPath, err)
			continue
		}
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
		ruleName := fmt.Sprintf("%s_lint_test", strings.ReplaceAll(name, "-", "_"))
		r := rule.NewRule(lintRule, ruleName)
		r.SetPrivateAttr("filename", e.Name())
		r.SetAttr("size", "small")
		r.SetAttr("src", e.Name())
		rules = append(rules, r)
	}

	for _, r := range rules {
		if len(deps) > 0 {
			r.SetAttr("deps", deps)
		}
		result.Gen = append(result.Gen, r)
		result.Imports = append(result.Imports, r.PrivateAttr("filename"))
	}
	return
}

func NewLanguage() language.Language {
	return &Jsonnet{Verbose: true}
}
