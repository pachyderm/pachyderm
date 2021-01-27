// Package ppsconsts constains constants relevant to PPS that are used across
// Pachyderm. In particular, the pipeline spec repo is handled specially by PFS
// and Auth, and those implementations need to refer to its name without
// depending on any other part of PPS. This package contains that and related
// constants as a minimal dependency for PFS and auth
package ppsconsts

const (
	// SpecRepo contains every pipeline's PipelineInfo (in its own branch)
	SpecRepo = "__spec__"

	// SpecRepoDesc is the description applied to the spec repo.
	SpecRepoDesc = "PPS pipeline specs repo."

	// SpecFile is the file in every SpecRepo commit containing the PipelineInfo
	SpecFile = "spec"

	// PPSTokenKey is a key (in etcd) that maps to PPS's auth token.
	// This is the token that PPS uses to authorize spec writes.
	PPSTokenKey = "master_token"

	// SpoutMarkerBranch is the branch that spouts use for keeping track of spout marker files
	SpoutMarkerBranch = "marker"
)
