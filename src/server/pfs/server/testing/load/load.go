package load

import (
	units "github.com/docker/go-units"
)

func DefaultCommitsSpec() *CommitsSpec {
	return &CommitsSpec{
		Count:           5,
		OperationsSpecs: []*OperationsSpec{defaultOperationsSpec()},
		ValidatorSpec:   &ValidatorSpec{},
	}
}

func defaultOperationsSpec() *OperationsSpec {
	return &OperationsSpec{
		Count: 5,
		FuzzOperationSpecs: []*FuzzOperationSpec{
			&FuzzOperationSpec{
				OperationSpec: &OperationSpec{
					&PutFileSpec{
						FilesSpec: defaultFilesSpec(),
					},
				},
				Prob: 1.0,
			},
		},
	}
}

func defaultFilesSpec() *FilesSpec {
	return &FilesSpec{
		Count: 5,
		FuzzFileSpecs: []*FuzzFileSpec{
			&FuzzFileSpec{
				FileSpec: &FileSpec{
					RandomFileSpec: &RandomFileSpec{
						FuzzSizeSpecs: defaultFuzzSizeSpecs(),
					},
				},
				Prob: 1.0,
			},
		},
	}
}

func defaultFuzzSizeSpecs() []*FuzzSizeSpec {
	return []*FuzzSizeSpec{
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 1 * units.KB,
				Max: 10 * units.KB,
			},
			Prob: 0.3,
		},
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 10 * units.KB,
				Max: 100 * units.KB,
			},
			Prob: 0.3,
		},
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 1 * units.MB,
				Max: 10 * units.MB,
			},
			Prob: 0.3,
		},
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 10 * units.MB,
				Max: 100 * units.MB,
			},
			Prob: 0.1,
		},
	}
}
