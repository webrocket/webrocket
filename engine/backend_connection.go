package engine

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

// backendConnection implements a wrapper for the TCP connection providing
// some concurrency tricks.
type backendConnection struct {
	// The underlaying connection.
	conn net.Conn
	// Internal semaphore.
	mtx sync.Mutex
}

// Internal constructor
// -----------------------------------------------------------------------------

// newBackendConnection wrapps the given connection into a new backend connection
// object.
//
// conn     - The connection to be wrapped.
//
// Returns a new backend connection.
func newBackendConnection(conn net.Conn) *backendConnection {
	return &backendConnection{conn: conn}
}

// Exported
// -----------------------------------------------------------------------------

// Recv receives data from the underlaying connection and maps it to
// the backend request structure. If there's no data to read it will block
// until new data appears.
//
// Returns read request or an error if something went wrong.
func (c *backendConnection) Recv() (req *backendRequest, err error) {
	var msg = [][]byte{}
	var buf = bufio.NewReader(c.conn)
	var possibleEom = false
	for {
		chunk, err := buf.ReadSlice('\n')
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
		msg = append(msg[:], chunk[:len(chunk)-1])
	}
	if len(msg) < 1 {
		err = errors.New("bad request")
		return
	}
	var aid, cmd []byte
	var from int
	if len(msg) >= 3 && len(msg[1]) == 0 {
		// <<<
		// identity\n
		// \n
		// command\n
		// ...
		// >>>
		aid, cmd, from = msg[0], msg[2], 3
	} else {
		// <<<
		// command\n
		// ...
		// >>>
		aid, cmd, from = nil, msg[0], 1
	}
	req = newBackendRequest(c, aid, cmd, msg[from:])
	return
}

// Send packs the command and frames together and sends it to the client.
//
// cmd    - The command to be sent.
// frames - The frames to be sent.
//
// Returns an error if something went wrong.
func (c *backendConnection) Send(cmd string, frames ...string) (err error) {
	payload := cmd + "\n"
	payload += strings.Join(frames, "\n")
	payload += "\n\r\n\r\n"
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.conn != nil {
		_, err = c.conn.Write([]byte(payload))
	}
	return
}

func (c *backendConnection) SetDeadline(ddl time.Time) {
	c.conn.SetDeadline(ddl)
}

// IsAlive returns whether the underlaying connection is alive or not.
func (c *backendConnection) IsAlive() bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.conn != nil
}

// Kill closes the underlaying connection.
func (c *backendConnection) Kill() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
