package lion

import "io"

type readPuller struct {
	reader       io.Reader
	unmarshaller Unmarshaller
}

func newReadPuller(reader io.Reader, unmarshaller Unmarshaller) *readPuller {
	return &readPuller{
		reader,
		unmarshaller,
	}
}

func (r *readPuller) Pull() (*EncodedEntry, error) {
	encodedEntry := &EncodedEntry{}
	if err := r.unmarshaller.Unmarshal(r.reader, encodedEntry); err != nil {
		return nil, err
	}
	return encodedEntry, nil
}
