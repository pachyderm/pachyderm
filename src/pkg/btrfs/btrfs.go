package btrfs

type API interface {
	//PropertySetReadonly(path string, readonly bool) error
	//PropertyGetReadonly(path string) (bool, error)
	SubvolumeCreate(path string) error
	//SubvolumeSnapshot(src string, dest string, readonly bool) error
	SubvolumeSnapshot(src string, dest string) error
}

func NewExecAPI() API {
	return newExecAPI()
}
