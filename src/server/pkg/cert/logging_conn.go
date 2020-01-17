// logging_conn.go defines LoggingListener, an implementation of net.Listener
// that allows tests to read the text transported over a client-server
// connection, and confirm, in the case of a TLS connection, that nothing
// sensitive was transported in plain text.

package cert

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

type loggingPipeAddr struct{}

func (o loggingPipeAddr) Network() string {
	return "logging-pipe"
}

func (o loggingPipeAddr) String() string {
	return "logging-pipe"
}

// loggingPipe is a struct containing two buffers, which log all of the traffic
// passing through LoggingPipe in each direction
type loggingPipe struct {
	ClientToServerBuf, ServerToClientBuf bytes.Buffer

	clientReader, serverReader io.Reader
	clientWriter, serverWriter io.WriteCloser
}

// newLoggingPipe initializes a loggingPipe
func newLoggingPipe() *loggingPipe {
	p := &loggingPipe{}
	p.clientReader, p.clientWriter = io.Pipe()
	p.clientReader = io.TeeReader(p.clientReader, &p.ServerToClientBuf)
	p.serverReader, p.serverWriter = io.Pipe()
	p.serverReader = io.TeeReader(p.serverReader, &p.ClientToServerBuf)
	return p
}

// Close closes 'l' (no more reading/writing will be possible)
func (p *loggingPipe) Close() error {
	p.clientWriter.Close()
	p.serverWriter.Close()
	return nil
}

// clientConn returns a loggingConn at the oposite end of the loggingPipe as
// serverConn. There is no fundamental difference between the clientConn and
// the serverConn, as communication is full duplex, but distinguishing the two
// ends of the pipe as client and server, rather than e.g. left and right,
// hopefully makes the calling code easier to read
func (p *loggingPipe) clientConn() *loggingConn {
	return &loggingConn{
		pipe: p,
		r:    p.clientReader,
		w:    p.serverWriter,
	}
}

// serverConn returns a loggingConn at the opposite end of the loggingPipe of
// clientConn (see the clientConn description for more information)
func (p *loggingPipe) serverConn() *loggingConn {
	return &loggingConn{
		pipe: p,
		r:    p.serverReader,
		w:    p.clientWriter,
	}
}

// loggingConn is an implementation of net.Conn that communicates with another
// party over a loggingPipe.
type loggingConn struct {
	// pipe is the loggingPipe over which this connection is communicating
	pipe *loggingPipe
	r    io.Reader
	w    io.WriteCloser
}

// Read implements the corresponding method of net.Conn
func (l *loggingConn) Read(b []byte) (n int, err error) {
	return l.r.Read(b)
}

// Write implements the corresponding method of net.Conn
func (l *loggingConn) Write(b []byte) (n int, err error) {
	return l.w.Write(b)
}

// Close implements the corresponding method of net.Conn
func (l *loggingConn) Close() error {
	return l.pipe.Close()
}

// LocalAddr implements the corresponding method of net.Conn
func (l *loggingConn) LocalAddr() net.Addr {
	return loggingPipeAddr{}
}

// RemoteAddr implements the corresponding method of net.Conn
func (l *loggingConn) RemoteAddr() net.Addr {
	return loggingPipeAddr{}
}

// SetDeadline implements the corresponding method of net.Conn
func (l *loggingConn) SetDeadline(t time.Time) error {
	return errors.New("not implemented")
}

// SetReadDeadline implements the corresponding method of net.Conn
func (l *loggingConn) SetReadDeadline(t time.Time) error {
	return errors.New("not implemented")
}

// SetWriteDeadline implements the corresponding method of net.Conn
func (l *loggingConn) SetWriteDeadline(t time.Time) error {
	return errors.New("not implemented")
}

// TestListener implements the net.Listener interface, returning loggingConns
type TestListener struct {
	// conn is the first (and last) non-nil connection returned from a call to
	// Accept()
	conn   *loggingConn
	connMu sync.Mutex

	// connCh provides connections (or nil) to Accept()
	connCh chan net.Conn
}

// NewTestListener initializes and returns a new TestListener. To create
// a new connection that that this Listener will serve on, call Dial(). To see
// the logged communication over that connection's pipe, see ClientToServerLog
// and ServerToClientLog
func NewTestListener() *TestListener {
	return &TestListener{
		connCh: make(chan net.Conn),
	}
}

// Dial initializes a new connection and releases a blocked call to Accept()
func (l *TestListener) Dial(context.Context, string, string) (net.Conn, error) {
	l.connMu.Lock()
	defer l.connMu.Unlock()
	if l.conn != nil {
		return nil, errors.New("Dial() has already been called on this TestListener")
	}

	// Initialize logging pipe
	p := newLoggingPipe()
	l.conn = p.serverConn()

	// send serverConn to Accept() and close l.connCh (so future callers to
	// Accept() get nothing)
	l.connCh <- p.serverConn()
	close(l.connCh)
	return p.clientConn(), nil
}

// ClientToServerLog returns the log of client -> server communication over the
// first (and only) connection spawned by this listener
func (l *TestListener) ClientToServerLog() []byte {
	return l.conn.pipe.ClientToServerBuf.Bytes()
}

// ServerToClientLog the log of server -> client communication over the first
// (and only) connection spawned by this listener
func (l *TestListener) ServerToClientLog() []byte {
	return l.conn.pipe.ServerToClientBuf.Bytes()
}

// Accept implements the corresponding method of net.Listener for
// TestListener
func (l *TestListener) Accept() (net.Conn, error) {
	conn := <-l.connCh
	if conn == nil {
		return nil, errors.New("Accept() has already been called on this TestListener")
	}
	return conn, nil
}

// Close implements the corresponding method of net.Listener for
// TestListener. Any blocked Accept operations will be unblocked and return
// errors.
func (l *TestListener) Close() error {
	l.connMu.Lock()
	defer l.connMu.Unlock()
	c := <-l.connCh
	if c != nil {
		close(l.connCh)
	}
	return nil
}

// Addr implements the corresponding method of net.Listener for
// TestListener
func (l *TestListener) Addr() net.Addr {
	return loggingPipeAddr{}
}
