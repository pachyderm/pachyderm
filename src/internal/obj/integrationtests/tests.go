package integrationtests

// BackendType is used to tell the tests which backend is being tested so some
// testing can be skipped for backends that do not support certain behavior.
type BackendType string

const (
	AmazonBackend BackendType = "Amazon"
	ECSBackend    BackendType = "ECS"
	GoogleBackend BackendType = "Google"
)

// ClientType is used to tell the tests which client is being tested so some testing
// can be skipped for clients that do not support certain behavior.
type ClientType string
