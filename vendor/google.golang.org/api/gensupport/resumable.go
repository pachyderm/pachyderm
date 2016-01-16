// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gensupport

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const (
	// statusResumeIncomplete is the code returned by the Google uploader when the transfer is not yet complete.
	statusResumeIncomplete = 308

	// uploadPause determines the delay between failed upload attempts
	uploadPause = 1 * time.Second
)

// ResumableUpload is used by the generated APIs to provide resumable uploads.
// It is not used by developers directly.
type ResumableUpload struct {
	Client *http.Client
	// URI is the resumable resource destination provided by the server after specifying "&uploadType=resumable".
	URI       string
	UserAgent string // User-Agent for header of the request
	// Media is the object being uploaded.
	Media io.ReaderAt
	// MediaType defines the media type, e.g. "image/jpeg".
	MediaType string
	// ContentLength is the full size of the object being uploaded.
	ContentLength int64

	mu       sync.Mutex // guards progress
	progress int64      // number of bytes uploaded so far

	// Callback is an optional function that will be periodically called with the cumulative number of bytes uploaded.
	Callback func(int64)
}

var (
	// rangeRE matches the transfer status response from the server. $1 is the last byte index uploaded.
	rangeRE = regexp.MustCompile(`^bytes=0\-(\d+)$`)
	// chunkSize is the size of the chunks created during a resumable upload and should be a power of two.
	// 1<<18 is the minimum size supported by the Google uploader, and there is no maximum.
	chunkSize int64 = 1 << 23
)

// Progress returns the number of bytes uploaded at this point.
func (rx *ResumableUpload) Progress() int64 {
	rx.mu.Lock()
	defer rx.mu.Unlock()
	return rx.progress
}

func (rx *ResumableUpload) transferStatus(ctx context.Context) (int64, *http.Response, error) {
	req, _ := http.NewRequest("POST", rx.URI, nil)
	req.ContentLength = 0
	req.Header.Set("User-Agent", rx.UserAgent)
	req.Header.Set("Content-Range", fmt.Sprintf("bytes */%v", rx.ContentLength))
	res, err := ctxhttp.Do(ctx, rx.Client, req)
	if err != nil || res.StatusCode != statusResumeIncomplete {
		return 0, res, err
	}
	var start int64
	if m := rangeRE.FindStringSubmatch(res.Header.Get("Range")); len(m) == 2 {
		start, err = strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("unable to parse range size %v", m[1])
		}
		start += 1 // Start at the next byte
	}
	return start, res, nil
}

func (rx *ResumableUpload) transferChunks(ctx context.Context) (*http.Response, error) {
	start, res, err := rx.transferStatus(ctx)
	if err != nil || res.StatusCode != statusResumeIncomplete {
		if err == context.Canceled {
			return &http.Response{StatusCode: http.StatusRequestTimeout}, err
		}
		return res, err
	}

	for {
		select { // Check for cancellation
		case <-ctx.Done():
			res.StatusCode = http.StatusRequestTimeout
			return res, ctx.Err()
		default:
		}
		reqSize := rx.ContentLength - start
		if reqSize > chunkSize {
			reqSize = chunkSize
		}
		r := io.NewSectionReader(rx.Media, start, reqSize)
		req, _ := http.NewRequest("POST", rx.URI, r)
		req.ContentLength = reqSize
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %v-%v/%v", start, start+reqSize-1, rx.ContentLength))
		req.Header.Set("Content-Type", rx.MediaType)
		req.Header.Set("User-Agent", rx.UserAgent)
		res, err = ctxhttp.Do(ctx, rx.Client, req)
		start += reqSize
		if err == nil && (res.StatusCode == statusResumeIncomplete || res.StatusCode == http.StatusOK) {
			rx.mu.Lock()
			rx.progress = start // keep track of number of bytes sent so far
			rx.mu.Unlock()
			if rx.Callback != nil {
				rx.Callback(start)
			}
		}
		if err != nil || res.StatusCode != statusResumeIncomplete {
			break
		}
	}
	return res, err
}

var sleep = time.Sleep // override in unit tests

// Upload starts the process of a resumable upload with a cancellable context.
// It retries indefinitely (with a pause of uploadPause between attempts) until cancelled.
// It is called from the auto-generated API code and is not visible to the user.
// rx is private to the auto-generated API code.
func (rx *ResumableUpload) Upload(ctx context.Context) (*http.Response, error) {
	var res *http.Response
	var err error
	for {
		res, err = rx.transferChunks(ctx)
		if err != nil || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
			return res, err
		}
		select { // Check for cancellation
		case <-ctx.Done():
			res.StatusCode = http.StatusRequestTimeout
			return res, ctx.Err()
		default:
		}
		sleep(uploadPause)
	}
	return res, err
}
