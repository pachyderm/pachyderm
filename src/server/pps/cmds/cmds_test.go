// TODO(msteffen) Add tests for:
//
// - restart-datum
// - stop-job
// - delete-job
//
// - inspect-job
// - list-job
//
// - create-pipeline
// - create-pipeline --push-images (re-enable existing test)
// - update-pipeline
// - delete-pipeline
//
// - inspect-pipeline
// - list-pipeline
//
// - start-pipeline
// - stop-pipeline
// - run-pipeline
//
// - list-datum
// - inspect-datum
// - get-logs

package cmds

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

const badJSON1 = `
{
"356weryt

}
`

const badJSON2 = `
{
    "a": 1,
    "b": [23,4,4,64,56,36,7456,7],
    "c": {a,f,g,h,j,j},
    "d": 3452.36456,
}
`

func TestJSONSyntaxErrorsReportedCreatePipeline(t *testing.T) {
	require.NoError(t, tu.BashCmd(`
		echo -n '{{.badJSON1}}' \
		  | ( pachctl create-pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"

		echo -n '{{.badJSON2}}' \
		  | ( pachctl create-pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"
		`,
		"badJSON1", badJSON1,
		"badJSON2", badJSON2,
	).Run())
}

func TestJSONSyntaxErrorsReportedUpdatePipeline(t *testing.T) {
	require.NoError(t, tu.BashCmd(`
		echo -n '{{.badJSON1}}' \
		  | ( pachctl update-pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"

		echo -n '{{.badJSON2}}' \
		  | ( pachctl update-pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"
		`,
		"badJSON1", badJSON1,
		"badJSON2", badJSON2,
	).Run())
}

// func TestPushImages(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}
// 	ioutil.WriteFile("test-push-images.json", []byte(`{
//   "pipeline": {
//     "name": "test_push_images"
//   },
//   "transform": {
//     "cmd": [ "true" ],
// 	"image": "test-job-shim"
//   }
// }`), 0644)
// 	os.Args = []string{"pachctl", "create-pipeline", "--push-images", "-f", "test-push-images.json"}
// 	require.NoError(t, rootCmd().Execute())
// }
