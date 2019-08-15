package common

// DatumID computes the id for a datum, this value is used in ListDatum and
// InspectDatum.
func DatumID(data []*Input) string {
	hash := sha256.New()
	for _, d := range data {
		hash.Write([]byte(d.FileInfo.File.Path))
		hash.Write(d.FileInfo.Hash)
	}
	// InputFileID is a single string id for the data from this input, it's used in logs and in
	// the statsTree
	return hex.EncodeToString(hash.Sum(nil))
}

// HashDatum computes and returns the hash of datum + pipeline, with a
// pipeline-specific prefix.
func HashDatum(pipelineName string, pipelineSalt string, data []*Input) string {
	hash := sha256.New()
	for _, datum := range data {
		hash.Write([]byte(datum.Name))
		hash.Write([]byte(datum.FileInfo.File.Path))
		hash.Write(datum.FileInfo.Hash)
	}

	hash.Write([]byte(pipelineName))
	hash.Write([]byte(pipelineSalt))

	return client.DatumTagPrefix(pipelineSalt) + hex.EncodeToString(hash.Sum(nil))
}

// HashDatum15 computes and returns the hash of datum + pipeline for version <= 1.5.0, with a
// pipeline-specific prefix.
func HashDatum15(pipelineInfo *pps.PipelineInfo, data []*Input) (string, error) {
	hash := sha256.New()
	for _, datum := range data {
		hash.Write([]byte(datum.Name))
		hash.Write([]byte(datum.FileInfo.File.Path))
		hash.Write(datum.FileInfo.Hash)
	}

	// We set env to nil because if env contains more than one elements,
	// since it's a map, the output of Marshal() can be non-deterministic.
	env := pipelineInfo.Transform.Env
	pipelineInfo.Transform.Env = nil
	defer func() {
		pipelineInfo.Transform.Env = env
	}()
	bytes, err := pipelineInfo.Transform.Marshal()
	if err != nil {
		return "", err
	}
	hash.Write(bytes)
	hash.Write([]byte(pipelineInfo.Pipeline.Name))
	hash.Write([]byte(pipelineInfo.ID))
	hash.Write([]byte(strconv.Itoa(int(pipelineInfo.Version))))

	// Note in 1.5.0 this function was called HashPipelineID, it's now called
	// HashPipelineName but it has the same implementation.
	return client.DatumTagPrefix(pipelineInfo.ID) + hex.EncodeToString(hash.Sum(nil)), nil
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}
