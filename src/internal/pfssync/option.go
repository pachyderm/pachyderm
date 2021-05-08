package pfssync

import "archive/tar"

// DownloadOption configures a download call.
type DownloadOption func(*downloadConfig)

// WithLazy configures the download call to lazily download files.
func WithLazy() DownloadOption {
	return func(dc *downloadConfig) {
		dc.lazy = true
	}
}

// WithEmpty configures the download call to just download the file info.
func WithEmpty() DownloadOption {
	return func(dc *downloadConfig) {
		dc.empty = true
	}
}

// WithHeaderCallback configures the download call to execute the callback for each tar file downloaded.
func WithHeaderCallback(cb func(*tar.Header) error) DownloadOption {
	return func(dc *downloadConfig) {
		dc.headerCallback = cb
	}
}
