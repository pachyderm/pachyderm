package protolog

import "io"

type readPuller struct {
	reader       io.Reader
	unmarshaller Unmarshaller
}

func newReadPuller(reader io.Reader, options ReadPullerOptions) *readPuller {
	readPuller := &readPuller{
		reader,
		options.Unmarshaller,
	}
	if readPuller.unmarshaller == nil {
		readPuller.unmarshaller = defaultUnmarshaller
	}
	return readPuller
}

func (r *readPuller) Pull() (*Entry, error) {
	entry := &Entry{}
	if err := r.unmarshaller.Unmarshal(r.reader, entry); err != nil {
		return nil, err
	}
	return entry, nil
}
