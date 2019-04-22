package index

import "github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"

func refsToOps(dataRefs []*chunk.DataRef) []*DataOp {
	var result []*DataOp
	for _, dataRef := range dataRefs {
		result = append(result, &DataOp{DataRef: dataRef})
	}
	return result
}

func opsToRefs(dataOps []*DataOp) []*chunk.DataRef {
	var result []*chunk.DataRef
	for _, dataOp := range dataOps {
		result = append(result, dataOp.DataRef)
	}
	return result
}
