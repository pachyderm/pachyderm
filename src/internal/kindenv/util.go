package kindenv

import (
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func readRunfile(rlocation string) ([]byte, error) {
	path, err := runfiles.Rlocation(rlocation)
	if err != nil {
		return nil, errors.Wrapf(err, "get runfile %v", rlocation)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read runfile %v", rlocation)
	}
	return content, nil
}
