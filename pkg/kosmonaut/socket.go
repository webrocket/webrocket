package kosmonaut

import (
	"bufio"
	uuid "github.com/nu7hatch/gouuid"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

// socket is a base struct for the Client and Worker.
type socket struct {
	// WebRocket backend's URL.
	URL *url.URL
	// The socket's identity.
	Identity string
	// Type of the socket.
	kind string
	// Internal semaphore.
	mtx sync.Mutex
}

// newSockets allocates memory for the new socket object of given kind.
//
// kind - Type of the socket.
// uri  - The backend endpoing URL to connect to.
//
// Returns configured socket or an error if something went wrong.
func newSocket(kind, uri string) (s *socket, err error) {
	s = &socket{kind: kind}
	s.URL, err = url.Parse(uri)
	return
}

// generateIdentity creates unique identity for the socket. Identity has
// the following format:
//
//     [socket-type]:[vhost]:[vhost-token]:[unique-id]
//
func (s *socket) generateIdentity() {
	id, _ := uuid.NewV4()
	parts := []string{s.kind, s.URL.Path, s.URL.User.Username(), id.String()}
	s.Identity = strings.Join(parts, ":")
}

// connect sets up new connection respecting given timeout.
//
// timeout - The request's maximum duration.
//
// Returns configured connection or an error if something went wrong.
func (s *socket) connect(timeout time.Duration) (conn net.Conn, err error) {
	if conn, err = net.DialTimeout("tcp", s.URL.Host, timeout); err != nil {
		return
	}
	s.generateIdentity()
	return
}

// recv reads the message from specified connection.
//
// c - The connection to read from.
//
// Returns list of received frames or an error if something went wrong.
func recv(c net.Conn) (frames []string, err error) {
	var buf = bufio.NewReader(c)
	var possibleEom = false
	for {
		var chunk []byte
		chunk, err = buf.ReadSlice('\n')
		if err != nil {
			break
		}
		if string(chunk) == "\r\n" {
			// Seems like it's end of the message...
			if possibleEom {
				// .. yeap, it is!
				break
			}
			possibleEom = true
			continue
		} else {
			possibleEom = false
		}
		frames = append(frames[:], string(chunk[:len(chunk)-1]))
	}
	return
}

// pack serializes given frames into backend protocol message.
//
// frames   - The frames to be serialized.
// Idnetity - The socket's identity. If not empty, it will be attached
//            to the message.
//
// Returns serialized message.
func pack(frames []string, identity string) (data []byte) {
	var res string
	if identity != "" {
		res = identity + "\n\n"
	}
	res += strings.Join(frames, "\n")
	res += "\n\r\n\r\n"
	return []byte(res)
}
