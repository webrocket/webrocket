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
	"log"
	"errors"
)

const (
	ReconnectDelay    = 1 * time.Second
	HeartbeatInterval = 500 * time.Millisecond
)

type Worker struct {
	*socket
	reconnectDelay time.Duration
	heartbeatIvl   time.Duration
	heartbeatAt    time.Time
	alive          bool
	conn           net.Conn
}

func NewWorker(uri string) (c *Worker, err error) {
	c = &Worker{
		alive:          false,
		reconnectDelay: ReconnectDelay,
		heartbeatIvl:   HeartbeatInterval,
	}
	c.socket, err = newSocket("dlr", uri)
	return
}

func (w *Worker) disconnect() {
	if w.conn == nil {
		return
	}
	w.conn.Close()
	w.conn = nil
}

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

func (w *Worker) send(frames []string, identity string) (err error) {
	if w.conn == nil {
		err = errors.New("not connected")
		return
	}
	packet := pack(frames, identity)
	_, err = w.conn.Write(packet)
	return
}

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
			w.send([]string{"QT"}, "")
			w.disconnect()
			break
		}
		ddl = time.Now().Add(w.heartbeatIvl * 2 + time.Second)
		w.conn.SetDeadline(ddl)
		if rawmsg, err = recv(w.conn); err != nil {
			println(err.Error())
			goto reconnect
		}
		if len(rawmsg) < 1 {
			continue
		}
		switch rawmsg[0] {
		case "HB":
			// Nothing to do...
			println("<-- HB")
		case "QT":
			// The endpoint is dead, we need to connect to another one.
			w.reconnect()
		case "TR":
			// Trigger the event.
			if msg, err := parseMessage(rawmsg[1:], w); err != nil {
				log.Printf("%v\n", err.Error())
			} else {
				ex <- msg
			}
		case "ER":
			// Notify about the error.
			e := parseError(rawmsg[1:])
			log.Printf(e.Error())
			if e.Code == EUnauthorized {
				break
			}
		}
		if time.Now().After(w.heartbeatAt) {
			println("--> HB")
			w.send([]string{"HB"}, "")
			w.heartbeatAt = time.Now().Add(w.heartbeatIvl)
		}
		continue
	}
}

func (w *Worker) Run() <-chan *Message {
	w.alive = true
	ex := make(chan *Message)
	go w.run(ex)
	return ex
}

func (w *Worker) Stop() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.alive = false
}

func (w *Worker) IsRunning() bool {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.alive
}