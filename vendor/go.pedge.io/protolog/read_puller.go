package protolog

import "io"

type readPuller struct {
	reader       io.Reader
	decoder      Decoder
	unmarshaller Unmarshaller
}

func newReadPuller(reader io.Reader, decoder Decoder, options ReadPullerOptions) *readPuller {
	readPuller := &readPuller{
		reader,
		decoder,
		options.Unmarshaller,
	}
	if readPuller.unmarshaller == nil {
		readPuller.unmarshaller = defaultUnmarshaller
	}
	return readPuller
}

func (r *readPuller) Pull() (*Entry, error) {
	data, err := r.decoder.Decode(r.reader)
	if err != nil {
		return nil, err
	}
	entry := &Entry{}
	if err := r.unmarshaller.Unmarshal(data, entry); err != nil {
		return nil, err
	}
	return entry, nil
}
