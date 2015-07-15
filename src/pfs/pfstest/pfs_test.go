package pfstest

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

func TestBasic(t *testing.T) {}

func test(
	t *testing.T,
	driver drive.Driver,
	f func(t *testing.T, apiClient pfs.ApiClient),
) {

}
