package btrfs

import "github.com/pachyderm/pachyderm/src/pkg/executil"

type execAPI struct{}

func newExecAPI() *execAPI {
	return &execAPI{}
}

//func (a *execAPI) PropertySetReadonly(path string, readonly bool) error {
//return execPropertySetReadonly(path, readonly)
//}

//func (a *execAPI) PropertyGetReadonly(path string) (bool, error) {
//return execPropertyGetReadonly(path)
//}

func (a *execAPI) SubvolumeCreate(path string) error {
	return execSubvolumeCreate(path)
}

//func (a *execAPI) SubvolumeSnapshot(src string, dest string, readonly bool) error {
//return execSubvolumeSnapshot(src, dest, readonly)
//}

func (a *execAPI) SubvolumeSnapshot(src string, dest string) error {
	return execSubvolumeSnapshot(src, dest)
}

//func execPropertySetReadonly(path string, readonly bool) error {
//readonlyString := "false"
//if readonly {
//readonlyString = "true"
//}
//return executil.Run("btrfs", "property", "set", path, "ro", readonlyString)
//}

//func execPropertyGetReadonly(path string) (bool, error) {
//reader, err := executil.RunStdout("btrfs", "property", "get", path)
//if err != nil {
//return false, err
//}
//scanner := bufio.NewScanner(reader)
//for scanner.Scan() {
//text := scanner.Text()
//if strings.Contains(text, "ro=true") {
//return true, nil
//}
//if strings.Contains(text, "ro=false") {
//return false, nil
//}
//}
//if err := scanner.Err(); err != nil {
//return false, err
//}
//return false, fmt.Errorf("did not find ro=true or ro=false for %s", path)
//}

func execSubvolumeCreate(path string) error {
	return executil.Run("btrfs", "subvolume", "create", path)
}

//func execSubvolumeSnapshot(src string, dest string, readonly bool) error {
func execSubvolumeSnapshot(src string, dest string) error {
	//if readonly {
	//return executil.Run("btrfs", "subvolume", "snapshot", "-r", src, dest)
	//}
	return executil.Run("btrfs", "subvolume", "snapshot", src, dest)
}
