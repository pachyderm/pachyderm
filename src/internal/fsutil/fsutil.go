package fsutil

import (
	"fmt"
	"io/fs"
	"sort"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type fileInfoWithFullName struct {
	fs.FileInfo
	name string
}

func (i *fileInfoWithFullName) Name() string {
	return i.name
}
func (i *fileInfoWithFullName) String() string {
	uid, _ := FileUID(i)
	gid, _ := FileGID(i)
	return fmt.Sprintf("uid:%d gid:%d %v", uid, gid, fs.FormatFileInfo(i))
}

// Find works like the UNIX "find" command, on an fs.FS.
func Find(f fs.FS) ([]fs.FileInfo, error) {
	var result []fs.FileInfo
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := fs.Stat(f, path)
		if err != nil {
			return errors.Wrapf(err, "stat %v", path)
		}
		result = append(result, &fileInfoWithFullName{name: path, FileInfo: info})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}
