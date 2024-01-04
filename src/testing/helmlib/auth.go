package helmlib

/**
auth.go is a utility library that assists with the process of deploying
Pachyderm into a Kubernetes cluster using its helm chart. Most of that process
is implemented in deploy.go, but this file assists with auth-related operations
necessary to that goal.

Currently, this code is only used as part of the 'pachdev' CLI tool (which is
officially the local testing tool of Pachyderm, and is also used, at least, by
our pachyderm-sdk python tests.

Like deploy.go, this library has a lot of copied code (see the comments at the
top of that file), but in this case, the code copied from
src/internal/testutil/auth.go, which is a Go library that assists
'minikubetestenv' in the same way that this library assists 'deploy.go'. As with
'deploy.go'/'minikubetestenv', This should eventually be unified with
testutil/auth.go. This was not done during 'pachdev' development for the same
reasons as 'deploy.go'--widespread use of '*testing.T' in 'testutil/auth.go' and
limiting the scope of developing 'pachdev'.
**/

const (
	// RootToken is the hard-coded admin token used on all activated test clusters
	RootToken = "iamroot"
)
