// Package ppsconsts constains constants relevant to PPS that are used across
// Pachyderm. In particular, the pipeline spec branch is handled specially by
// PFS, and PFS needs to refer to its name without depending on any other part
// of PPS. This package contains that and related constants as a minimal
// dependency for PFS.
package ppsconsts

const (
	// SpecBranch is a branch in every pipeline's output repo that contains the
	// pipeline's PipelineInfo
	SpecBranch = "spec"

	// SpecFile is the file in every SpecBranch commit containing the PipelineInfo
	SpecFile = "spec"
)
