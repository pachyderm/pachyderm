package cmdutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/spf13/cobra"
)

func TestGetStringFlag(t *testing.T) {
	// Returns a zero-value for an empty flag
	emptyFlag := BatchArgs{[]string{}, map[string][]string{}}
	require.Equal(t, emptyFlag.GetStringFlag("foo"), "")

	// Returns the specified value for a flag given once
	singleFlag := BatchArgs{[]string{}, map[string][]string{}}
	require.Equal(t, singleFlag.GetStringFlag("foo"), "")

	// Returns the last value for a flag given multiple times
	multiFlag := BatchArgs{[]string{}, map[string][]string{}}
	require.Equal(t, multiFlag.GetStringFlag("foo"), "")
}

func TestPartitionBatchArgs(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("value", "v", "", "A value.")
	cmd.Flags().BoolP("bool", "b", false, "A boolean.")

	// Partitions a simple array of positionals
	args := []string{"foo", "bar", "baz"}
	expected := [][]string{{"foo"}, {"bar"}, {"baz"}}
    actual, err := partitionBatchArgs(args, 1, cmd)
	require.NoError(t, err)
	require.Equal(t, actual, expected)

	// Partitions with value flags
	args = []string{"--value", "val", "foo", "--value", "val2", "bar", "-v", "val3"}
	expected = [][]string{{"--value", "val", "foo", "--value", "val2"}, {"bar", "-v", "val3"}}
    actual, err = partitionBatchArgs(args, 1, cmd)
	require.NoError(t, err)
	require.Equal(t, actual, expected)

	// Handles it ok if there is a trailing value flag with no value
	args = []string{"foo", "--value"}
	expected = [][]string{{"foo", "--value"}}
    actual, err = partitionBatchArgs(args, 1, cmd)
	require.NoError(t, err)
	require.Equal(t, actual, expected)

	args = []string{"foo", "-v"}
	expected = [][]string{{"foo", "-v"}}
    actual, err = partitionBatchArgs(args, 1, cmd)
	require.NoError(t, err)
	require.Equal(t, actual, expected)

	// Partitions with boolean flags
	args = []string{"--bool", "foo", "-b", "bar", "--bool"}
	expected = [][]string{{"--bool", "foo", "-b"}, {"bar", "--bool"}}
    actual, err = partitionBatchArgs(args, 1, cmd)
	require.NoError(t, err)
	require.Equal(t, actual, expected)

	// Partitions larger sets of positionals
	args = []string{"foo", "bar", "baz", "floop", "bloop", "hunter2"}
	expected = [][]string{{"foo", "bar", "baz"}, {"floop", "bloop", "hunter2"}}
    actual, err = partitionBatchArgs(args, 3, cmd)
	require.NoError(t, err)
	require.Equal(t, actual, expected)

	// Errors if the wrong number of positionals was passed in
	args = []string{"foo", "bar", "baz", "floop", "bloop"}
    actual, err = partitionBatchArgs(args, 3, cmd)
	require.YesError(t, err)

	args = []string{"foo", "bar", "baz", "floop"}
    actual, err = partitionBatchArgs(args, 3, cmd)
	require.YesError(t, err)
}
