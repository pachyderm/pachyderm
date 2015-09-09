package clone

import (
	"io/ioutil"
	"testing"

	"go.pedge.io/protolog/logrus"

	"github.com/pachyderm/pachyderm/src/pkg/require"
)

func init() {
	logrus.Register()
}

func TestBasic(t *testing.T) {
	dirPath, err := ioutil.TempDir("", "pachyderm")
	require.NoError(t, err)
	err = GithubClone(
		dirPath,
		"pachyderm",
		"pachyderm",
		"master",
		"11eb4c6e1945beb1e6ce3e878ed2cb6c24ee8bf1",
		"",
	)
	require.NoError(t, err)
	//data, err := ioutil.ReadFile(filepath.Join(dirPath, "README.md"))
	//require.NoError(t, err)
	//require.Equal(
	//t,
	//`pfs
	//===

	//The Pachyderm Filesystem`,
	//string(data),
	//)
}
