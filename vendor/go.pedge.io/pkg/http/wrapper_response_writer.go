package pkghttp

import "net/http"

type wrapperResponseWriter struct {
	http.ResponseWriter
	statusCode int
	writeError error
}

func newWrapperResponseWriter(responseWriter http.ResponseWriter) *wrapperResponseWriter {
	return &wrapperResponseWriter{responseWriter, 0, nil}
}

func (w *wrapperResponseWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.writeError = err
	return n, err
}

func (w *wrapperResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
