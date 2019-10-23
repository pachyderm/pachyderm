package cmds

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func run(t *testing.T, cmd string) error {
	t.Helper()

	tmpfile, err := ioutil.TempFile("", "test-pach-config-*.json")
	require.NoError(t, err)

	// remove the empty file so that a config can be generated
	require.NoError(t, os.Remove(tmpfile.Name()))

	// remove the config file when done
	defer os.Remove(tmpfile.Name())

	return tu.BashCmd(`
		export PACH_CONFIG={{.config}}
		{{.cmd}}
		`,
		"config", tmpfile.Name(),
		"cmd", cmd,
	).Run()
}

func TestInvalidEnvValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.YesError(t, run(t, `
		export PACH_CONTEXT=foobar
		pachctl config get active-context
	`))
}

func TestEnvValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo --overwrite
		echo '{}' | pachctl config set context bar --overwrite
		pachctl config set active-context bar
		export PACH_CONTEXT=foo
		pachctl config get active-context | match foo
	`))
}

func TestMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.NoError(t, run(t, `
		pachctl config get metrics | match true
		pachctl config set metrics false
		pachctl config get metrics | match false
		pachctl config set metrics true
		pachctl config get metrics | match true
	`))
}

func TestActiveContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.YesError(t, run(t, `
		pachctl config set active-context foo 2>&1 | match "context does not exist: foo"
	`))

	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo --overwrite
		pachctl config set active-context foo
		pachctl config get active-context | match "foo"
	`))
}

func TestSetContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.YesError(t, run(t, `
		echo 'malformed_json' | pachctl config set context foo
	`))

	require.YesError(t, run(t, `
		echo '{}' | pachctl config set context foo
		echo '{}' | pachctl config set context foo
	`))

	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo
		echo '{"pachd_address": "foobar:9000"}' | pachctl config set context foo --overwrite
		pachctl config get context foo | match '"pachd_address": "grpc://foobar:9000"'
	`))
}

func TestUpdateContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.YesError(t, run(t, `
		pachctl config update context foo --pachd-address=bar
	`))

	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo
		pachctl config update context foo --pachd-address="foobar:9000"
		pachctl config get context foo | match '"pachd_address": "grpc://foobar:9000"'
		pachctl config update context foo --pachd-address=""
		pachctl config get context foo | match -v pachd_address
	`))
}

func TestDeleteContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.YesError(t, run(t, `
		pachctl config delete context foo
	`))

	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo
		pachctl config delete context foo
	`))
}

func TestConfigListContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo
		echo '{}' | pachctl config set context bar
		pachctl config list context | match -v "\*	bar"
		pachctl config list context | match "	bar"
		pachctl config list context | match "	foo"

		pachctl config set active-context bar
		pachctl config list context | match "\*	bar"
		pachctl config list context | match "	foo"
	`))
}
