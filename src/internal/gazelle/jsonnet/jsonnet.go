// Package jsonnet is a Gazelle plugin that generates rules when it sees jsonnet files in the
// repository.
package jsonnet

import (
	"github.com/bazelbuild/bazel-gazelle/language"
)

type Jsonnet struct {
	language.BaseLang // Stubs to implement everything we don't care about.
}

func NewLanguage() language.Language {
	return &Jsonnet{}
}
