// Copyright (C) 2011 by Krzysztof Kowalik <chris@nu7hat.ch> and folks at Cubox
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package kosmonaut

import (
	"net"
	"time"
	"errors"
)

const (
	// The number of seconds to wait before next reconnect try.
	ReconnectDelay = 1 * time.Second
	// The number of milliseconds between the heartbeat messages.
	HeartbeatInterval = 500 * time.Millisecond
)

// Worker is a SUB socket implementation which handles asynchronous
// communication between backend application and WebRocket backend endpoint.
//
// Worker is used to listen for incoming messages and handle them in
// appropriate way.
type Worker struct {
	*socket
	// The delay between reconnect tries.
	reconnectDelay time.Duration
	// The heartbeat interval.
	heartbeatIvl time.Duration
	// The time of the next heartbeat message.
	heartbeatAt time.Time
	// The socket status - whether is running or not 
	alive bool
	// Underlaying TCP connection.
	conn net.Conn
}

// NewWorker allocates memory and preconfigures the SUB client.
//
// uri - The WebRocket backend's URL to connect to.
//
// Returns a Worker instance or an error if something went wrong.
func NewWorker(uri string) (c *Worker, err error) {
	c = &Worker{
		alive:          false,
		reconnectDelay: ReconnectDelay,
		heartbeatIvl:   HeartbeatInterval,
	}
	c.socket, err = newSocket("dlr", uri)
	return
}

// disconnect closes and cleans up the active connection. 
func (w *Worker) disconnect() {
	if w.conn == nil {
		return
	}
	w.conn.Close()
	w.conn = nil
}

// reconnect cleans up existing connection and tries to connect again.
//
// Returns an error if something went wrong.
func (w *Worker) reconnect() (err error) {
	w.disconnect()
	w.conn, err = w.connect(w.heartbeatIvl * 2 + 1)
	if err != nil {
		return
	}
	ddl := time.Now().Add(w.heartbeatIvl * 2)
	w.conn.SetWriteDeadline(ddl)
	w.send([]string{"RD"}, w.Identity)
	return
}

// Send packs and writes given data to the active connection.
//
// frames   - The frames to be packed and sent.
// identity - Worker's identity to be attached to the message.
//
// Returns an error if something went wrong.
func (w *Worker) send(frames []string, identity string) (err error) {
	if w.conn == nil {
		err = errors.New("not connected")
		return
	}
	packet := pack(frames, identity)
	_, err = w.conn.Write(packet)
	return
}

// run contains a listener's event loop and processes all the events
// incoming from the server.
//
// ex - An exchange (output) channel.
//
func (w *Worker) run(ex chan *Message) {
	defer close(ex)
	var err error
reconnect:
	if err = w.reconnect(); err != nil {
		// Keep reconnecting...
		<-time.After(w.reconnectDelay)
		goto reconnect
	}
	for {
		var rawmsg []string
		var ddl time.Time
		if w.conn == nil {
			goto reconnect
		}
		if !w.IsRunning() {
			// Worker has been turned off!
			w.send([]string{"QT"}, "")
			w.disconnect()
			break
		}
		ddl = time.Now().Add(w.heartbeatIvl * 2 + time.Second)
		w.conn.SetDeadline(ddl)
		if rawmsg, err = recv(w.conn); err != nil {
			// Couldn't get the message, reconnecting...
			goto reconnect
		}
		if len(rawmsg) < 1 {
			continue
		}
		switch rawmsg[0] {
		case "HB":
			// Nothing to do...
		case "QT":
			// The endpoint is dead, we need to connect to another one.
			w.reconnect()
		case "TR":
			// Trigger the event.
			msg := parseMessage(rawmsg[1:], w)
			ex <- msg
		case "ER":
			// Notify about the error.
			e := parseError(rawmsg[1:])
			ex <- &Message{Error: e}
			if e.Code == EUnauthorized {
				return
			}
		}
		if time.Now().After(w.heartbeatAt) {
			// Send a hartbeat message if it's time...
			w.send([]string{"HB"}, "")
			w.heartbeatAt = time.Now().Add(w.heartbeatIvl)
		}
		continue
	}
}

// Run starts event loop of the worker and returns a channel with
// incoming messages.
func (w *Worker) Run() <-chan *Message {
	w.alive = true
	ex := make(chan *Message)
	go w.run(ex)
	return ex
}

// Stop terminates event loop of the worker.
func (w *Worker) Stop() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.alive = false
}

// IsRunning returns whether the worker is running or not.
func (w *Worker) IsRunning() bool {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.alive
}