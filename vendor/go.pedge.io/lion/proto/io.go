/*
Some of the code here is copied from https://github.com/matttproud/golang_protobuf_extensions/tree/master/pbutil.

This code is under the Apache 2.0 License that can be found at https://github.com/matttproud/golang_protobuf_extensions/blob/master/LICENSE.
*/

package protolion

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"

	"github.com/golang/protobuf/proto"
)

var errInvalidVarint = errors.New("invalid varint32 encountered")

// writeDelimited encodes and dumps a message to the provided writer prefixed
// with a 32-bit varint indicating the length of the encoded message, producing
// a length-delimited record stream, which can be used to chain together
// encoded messages of the same type together in a file.  It returns the total
// number of bytes written and any applicable error.  This is roughly
// equivalent to the companion Java API's MessageLite#writeDelimitedTo.
func writeDelimited(w io.Writer, m proto.Message, base64Encode bool, newline bool) (int, error) {
	buffer, err := proto.Marshal(m)
	if err != nil {
		return 0, err
	}
	if base64Encode {
		buffer = []byte(base64.StdEncoding.EncodeToString(buffer))
	}

	buf := make([]byte, binary.MaxVarintLen32)
	encodedLength := binary.PutUvarint(buf, uint64(len(buffer)))

	sync, err := w.Write(buf[:encodedLength])
	if err != nil {
		return sync, err
	}

	n, err := w.Write(buffer)
	sync += n
	if err != nil {
		return sync, err
	}
	if newline {
		n, err = w.Write([]byte{'\n'})
		sync += n
	}
	return sync, err
}

// readDelimited decodes a message from the provided length-delimited stream,
// where the length is encoded as 32-bit varint prefix to the message body.
// It returns the total number of bytes read and any applicable error.  This is
// roughly equivalent to the companion Java API's
// MessageLite#parseDelimitedFrom.  As per the reader contract, this function
// calls r.Read repeatedly as required until exactly one message including its
// prefix is read and decoded (or an error has occurred).  The function never
// reads more bytes from the stream than required.  The function never returns
// an error if a message has been read and decoded correctly, even if the end
// of the stream has been reached in doing so.  In that case, any subsequent
// calls return (0, io.EOF).
func readDelimited(r io.Reader, m proto.Message, base64Decode bool, newline bool) (int, error) {
	// Per AbstractParser#parsePartialDelimitedFrom with
	// CodedInputStream#readRawVarint32.
	headerBuf := make([]byte, binary.MaxVarintLen32)
	var bytesRead, varIntBytes int
	var messageLength uint64
	for varIntBytes == 0 { // i.e. no varint has been decoded yet.
		if bytesRead >= len(headerBuf) {
			return bytesRead, errInvalidVarint
		}
		// We have to read byte by byte here to avoid reading more bytes
		// than required. Each read byte is appended to what we have
		// read before.
		newBytesRead, err := r.Read(headerBuf[bytesRead : bytesRead+1])
		if newBytesRead == 0 {
			if err != nil {
				return bytesRead, err
			}
			// A Reader should not return (0, nil), but if it does,
			// it should be treated as no-op (according to the
			// Reader contract). So let's go on...
			continue
		}
		bytesRead += newBytesRead
		// Now present everything read so far to the varint decoder and
		// see if a varint can be decoded already.
		messageLength, varIntBytes = proto.DecodeVarint(headerBuf[:bytesRead])
	}

	messageBuf := make([]byte, messageLength)
	newBytesRead, err := io.ReadFull(r, messageBuf)
	bytesRead += newBytesRead
	if err != nil {
		return bytesRead, err
	}
	if base64Decode {
		messageBuf, err = base64.StdEncoding.DecodeString(string(messageBuf))
		if err != nil {
			return bytesRead, err
		}
	}
	if newline {
		newlineBuf := make([]byte, 1)
		n, err := io.ReadFull(r, newlineBuf)
		bytesRead += n
		if err != nil {
			return bytesRead, err
		}
	}
	return bytesRead, proto.Unmarshal(messageBuf, m)
}
