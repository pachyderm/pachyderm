package s2

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

var (
	// chunkValidator is a regexp for validating a chunk "header" in the
	// request body of a multi-chunk upload
	chunkValidator = regexp.MustCompile(`^([0-9a-fA-F]+);chunk-signature=([0-9a-fA-F]+)`)

	// InvalidChunk is an error returned when reading a multi-chunk object
	// upload that contains an invalid chunk header or body
	InvalidChunk = errors.New("invalid chunk")
)

// Reads a multi-chunk upload body
type chunkedReader struct {
	body      io.ReadCloser
	lastChunk []byte
	bufBody   *bufio.Reader

	signingKey    []byte
	lastSignature string
	timestamp     string
	date          string
	region        string
}

func newChunkedReader(body io.ReadCloser, signingKey []byte, seedSignature, timestamp, date, region string) *chunkedReader {
	return &chunkedReader{
		body:      body,
		lastChunk: nil,
		bufBody:   bufio.NewReader(body),

		signingKey:    signingKey,
		lastSignature: seedSignature,
		timestamp:     timestamp,
		date:          date,
		region:        region,
	}
}

func (c *chunkedReader) Read(p []byte) (n int, err error) {
	if c.lastChunk == nil {
		if err := c.readChunk(); err != nil {
			return 0, err
		}
	}

	n = copy(p, c.lastChunk)

	if n == len(c.lastChunk) {
		c.lastChunk = nil
	} else {
		c.lastChunk = c.lastChunk[n:]
	}

	return n, nil
}

func (c *chunkedReader) readChunk() error {
	// step 1: read the chunk header
	line, err := c.bufBody.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return err
		}
		return InvalidChunk
	}

	match := chunkValidator.FindStringSubmatch(line)
	if len(match) == 0 {
		return InvalidChunk
	}

	chunkLengthHexStr := match[1]
	chunkSignature := match[2]

	chunkLength, err := strconv.ParseUint(chunkLengthHexStr, 16, 32)
	if err != nil {
		return InvalidChunk
	}

	// step 2: read the chunk body
	chunk := make([]byte, chunkLength)
	_, err = io.ReadFull(c.bufBody, chunk)
	if err != nil {
		return InvalidChunk
	}

	// step 3: read the trailer
	trailer := make([]byte, 2)
	_, err = io.ReadFull(c.bufBody, trailer)
	if err != nil || trailer[0] != '\r' || trailer[1] != '\n' {
		return InvalidChunk
	}

	// step 4: construct the string to sign
	stringToSign := fmt.Sprintf(
		"AWS4-HMAC-SHA256-PAYLOAD\n%s\n%s/%s/s3/aws4_request\n%s\ne3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855\n%x",
		c.timestamp,
		c.date,
		c.region,
		c.lastSignature,
		sha256.Sum256(chunk),
	)

	// step 5: calculate & verify the signature
	signature := hmacSHA256(c.signingKey, stringToSign)
	if chunkSignature != fmt.Sprintf("%x", signature) {
		return InvalidChunk
	}

	c.lastChunk = chunk
	c.lastSignature = chunkSignature
	return nil
}

func (c *chunkedReader) Close() error {
	return c.body.Close()
}
