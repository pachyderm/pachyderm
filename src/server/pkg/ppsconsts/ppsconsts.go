// Package ppsconsts constains constants relevant to PPS that are used across
// Pachyderm. In particular, the pipeline spec repo is handled specially by PFS
// and Auth, and those implementations need to refer to its name without
// depending on any other part of PPS. This package contains that and related
// constants as a minimal dependency for PFS and auth
package ppsconsts

const (
	// SpecRepo contains every pipeline's PipelineInfo (in its own branch)
	SpecRepo = "spec"

	// SpecFile is the file in every SpecRepo commit containing the PipelineInfo
	SpecFile = "spec"
)
