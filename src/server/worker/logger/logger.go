package logger

type TaggedLogger interface {
	io.Writer

	func Logf(formatString string, args ...interface{})

  func WithJob(jobID string) TaggedLogger
	func WithData(data []*Input) TaggedLogger
	func WithUserCode() TaggedLogger
}

type taggedLogger struct {
	template     pps.LogMessage
	stderrLog    log.Logger
	marshaler    *jsonpb.Marshaler

	// Used for mirroring log statements to object storage
	putObjClient pfs.ObjectAPI_PutObjectClient
	objSize      int64
	msgCh        chan string
	buffer			 bytes.Buffer
	eg           errgroup.Group
}

// TODO: the whole object api interface here is bad, there are a few shortcomings:
//  - it's only used under the worker function when stats are enabled
//  - 'Close' is used to end the object, but it must be explicitly called
//  - the 'eg', 'putObjClient', 'msgCh', 'buffer', and 'objSize' don't play well
//      with cloned loggers.
// Abstract this into a separate object with a more explicit lifetime?

// If a pachClient is passed in, and stats are enabled on the pipeline, log
// statements will be chunked and written to the Object API on the given
// client.
func NewLogger(pipelineInfo *pps.PipelineInfo, pachClient *client.APIClient) (TaggedLogger, error) {
	result := &taggedLogger{
		template:  pps.LogMessage{
			PipelineName: pipelineInfo.Pipeline.Name,
			WorkerID:     os.Getenv(client.PPSPodNameEnv),
		},
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
		msgCh:     make(chan string, logBuffer),
	}
	result.stderrLog.SetOutput(os.Stderr)
	result.stderrLog.SetFlags(log.LstdFlags | log.Llongfile) // Log file/line

	if pachClient != nil && pipelineInfo.EnableStats {
		putObjClient, err := pachClient.ObjectAPIClient.PutObject(pachClient.Ctx())
		if err != nil {
			return nil, err
		}
		result.putObjClient = putObjClient
		result.eg.Go(func() error {
			for msg := range result.msgCh {
				for _, chunk := range grpcutil.Chunk([]byte(msg), grpcutil.MaxMsgSize/2) {
					if err := result.putObjClient.Send(&pfs.PutObjectRequest{
						Value: chunk,
					}); err != nil && err != io.EOF {
						return err
					}
				}
				result.objSize += int64(len(msg))
			}
			return nil
		})
	}
	return result, nil
}

func NewMasterLogger(pipelineInfo *pps.PipelineInfo) (TaggedLogger, error) {
	result, err := newLogger(pipelineInfo, nil) // master loggers don't log stats
	if err != nil {
		return nil, err
	}
	result.template.Master = true
	return result, nil
}

func (logger *taggedLogger) WithJob(jobID string) TtaggedLogger {
	result := logger.clone()
	result.template.JobID = jobID
	return result
}

func (logger *taggedLogger) WithData(data []*Input) TaggedLogger {
	result := logger.clone()

	// Add inputs' details to log metadata, so we can find these logs later
	result.template.Data = []pps.InputFile{}
	for _, d := range data {
		result.template.Data = append(result.template.Data, &pps.InputFile{
			Path: d.FileInfo.File.Path,
			Hash: d.FileInfo.Hash,
		})
	}

	// InputFileID is a single string id for the data from this input, it's used in logs and in
	// the statsTree
	result.template.DatumID = DatumID(data)
	return result
}

func (logger *taggedLogger) WithUserCode() TaggedLogger {
	result := logger.clone()
	result.template.User = true
	return result
}

func (logger *taggedLogger) clone() *taggedLogger {
	return &taggedLogger{
		template:     logger.template, // Copy struct
		stderrLog:    log.Logger{},
		marshaler:    &jsonpb.Marshaler{},
		putObjClient: logger.putObjClient,
		msgCh:        logger.msgCh,
	}
}

// Logf logs the line Sprintf(formatString, args...), but formatted as a json
// message and annotated with all of the metadata stored in 'loginfo'.
//
// Note: this is not thread-safe, as it modifies fields of 'logger.template'
func (logger *taggedLogger) Logf(formatString string, args ...interface{}) {
	logger.template.Message = fmt.Sprintf(formatString, args...)
	if ts, err := types.TimestampProto(time.Now()); err == nil {
		logger.template.Ts = ts
	} else {
		logger.stderrLog.Printf("could not generate logging timestamp: %s\n", err)
		return
	}
	msg, err := logger.marshaler.MarshalToString(&logger.template)
	if err != nil {
		logger.stderrLog.Printf("could not marshal %v for logging: %s\n", &logger.template, err)
		return
	}
	fmt.Println(msg)
	if logger.putObjClient != nil {
		logger.msgCh <- msg + "\n"
	}
}

// This is provided so that taggedLogger can be used as a io.Writer for stdout
// and stderr when running user code.
func (logger *taggedLogger) Write(p []byte) (_ int, retErr error) {
  // never errors
  logger.buffer.Write(p)
  r := bufio.NewReader(&logger.buffer)
  for {
    message, err := r.ReadString('\n')
    if err != nil {
      if err == io.EOF {
        logger.buffer.Write([]byte(message))
        return len(p), nil
      }
      // this shouldn't technically be possible to hit io.EOF should be
      // the only error bufio.Reader can return when using a buffer.
      return 0, fmt.Errorf("error ReadString: %v", err)
    }
    // We don't want to make this call as:
    // logger.Logf(message)
    // because if the message has format characters like %s in it those
    // will result in errors being logged.
    logger.Logf("%s", strings.TrimSuffix(message, "\n"))
  }
}

// Close flushes and closes the object storage client used to mirror log
// statements to object storage.  Returns a pointer to the generated pfs.Object
// as well as the total size of all written messages.
func (logger *taggedLogger) Close() (*pfs.Object, int64, error) {
	close(logger.msgCh)
	if logger.putObjClient != nil {
		if err := logger.eg.Wait(); err != nil {
			return nil, 0, err
		}
		object, err := logger.putObjClient.CloseAndRecv()
		// we set putObjClient to nil so that future calls to Logf won't send
		// msg down logger.msgCh as we've just closed that channel.
		logger.putObjClient = nil
		return object, logger.objSize, err
	}
	return nil, 0, nil
}
