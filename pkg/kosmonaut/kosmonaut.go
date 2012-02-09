package kosmonaut

import (
	"fmt"
	"net/url"
	"net"
	"time"
	"strings"
	"sync"
	"bufio"
	"strconv"
	"encoding/json"
	uuid "github.com/nu7hatch/gouuid"
)

// Error represents a WebRocket protocol error.
type Error struct {
	Status string
	Code   int
}

// Error returns stringified status code with message.
func (err *Error) Error() string {
	return fmt.Sprintf("%d %s", err.Code, err.Status)
}

// Possible status messages.
var statusMessages = map[int]string{
	400: "Bad Request",
    402: "Unauthorized",
    403: "Forbidden",
    451: "Invalid channel name",
    454: "Channel not found",
    597: "Internal error",
    598: "End of file",
}

// Timeout value for the client requests. 
const RequestTimeout = 5 * time.Second

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
func (s *socket) recv(c net.Conn) (frames []string, err error) {
	var buf = bufio.NewReader(c)
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
		frames = append(frames[:], string(chunk[:len(chunk)-1]))
	}
	return
}

// pack serializes given frames into backend protocol message.
//
// frames       - The frames to be serialized.
// withIdnetity - Whether the identity should be prepend to the message.
//
// Returns serialized message.
func (s *socket) pack(frames []string, withIdentity bool) (data []byte) {
	var res string
	if withIdentity {
		res += s.Identity + "\n\n"
	}
	res += strings.Join(frames, "\n")
	res += "\n\r\n\r\n"
	return []byte(res)
}

// Client is an implementation REQ-REP type socket which handles communication
// between backend application and WebRocket backend endpoint.
// 
// Client is used to synchronously request operations from the server. Synchronous
// operations are used to provide consistency for the backed generated events.
type Client struct {
	*socket
}

// NewCLient allocates memory ancc preconfigures the REQ client.
//
// uri - The WebRocket backend's URL to connect to.
//
// Returns client instance or an error if something went wrong.
func NewClient(uri string) (c *Client, err error) {
	c = &Client{}
	c.socket, err = newSocket("req", uri)
	return
}

// performRequest sends given payload to the server and waits for the response.
//
// payload - A message to be sent.
//
// Returns received data or an error if something went wrong.
func (c *Client) performRequest(payload []string) (data string, err error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	var conn net.Conn
	var response []string
	if conn, err = c.connect(RequestTimeout); err != nil {
		return
	}
	defer conn.Close()
	packet := c.pack(payload, true)
	conn.Write(packet)
	if response, err = c.recv(conn); err != nil {
		return
	}
	return c.parseResponse(response)
}

// parseResponse takses received frames and extracts data from it. If payload
// contains the error information then it returns appropriate error.
//
// frames - A message to be parsed.
//
// Returns data extracted from the message or an error if something went wrong,
// or frames contains error payload.
func (c *Client) parseResponse(frames []string) (data string, err error) {
	if len(frames) < 1 {
		goto unknownError
	}
	switch frames[0] {
	case "OK":
		return "", nil
	case "ER":
		if len(frames) < 2 {
			goto unknownError
		}
		code, _ := strconv.Atoi(frames[1])
		status, ok := statusMessages[code]
		if !ok {
			goto unknownError
		}
		return "", &Error{status, code}
	case "AT":
		if len(frames) < 2 {
			goto unknownError
		}
		token := frames[1]
		if len(token) != 128 {
			goto unknownError
		}
		return token, nil
	}
unknownError:
	return "", &Error{"Unknown server error", 0}
}

// Open opens specified channel. If channel already exists, then ok response
// will be received anyway. If channel name is starts with the `presence-`
// or `private-` prefix, then appropriate type of the channel will be created.
// 
// name - A name of the channel to be created.
// 
// Returns an error if something went wrong.
func (c *Client) OpenChannel(name string) (err error) {
	payload := []string{"OC", name}
	_, err = c.performRequest(payload)
	return
}

// Close closes specified channel. If channel doesn't exist then an error will
// be thrown.
//
// name - A name of the channel to be created.
// 
// Returns an error if something went wrong.
func (c *Client) CloseChannel(name string) (err error) {
	payload := []string{"CC", name}
	_, err = c.performRequest(payload)
	return
}

// Broadcast sends an event with attached data on the specified channel.
// 
// channel - A name of the channel to broadcast to.
// event   - A name of the event to be triggered.
// data    - The data attached to the event.
// 
// Examples
// 
//     c.Broadcast("room", "away", {"message" => "on the meeting"})
//     c.Broadcast("room". "message", {"content" => "Hello World!"})
//     c.Broadcast("room". "status", {"message" => "is saying hello!"})
// 
// Returns an error if something went wrong.
func (c *Client) Broadcast(channel, event string, data map[string]interface{}) (err error) {	
	var serialized []byte
	if serialized, err = json.Marshal(data); err != nil {
		return
	}
	payload := []string{"BC", channel, event, string(serialized)}
	_, err = c.performRequest(payload)
	return
}

// RequestSingleAccessToken sends a request to generate a single access token
// for given user with specified permissions.
// 
// uid        - An user defined unique ID.
// permission - A permissions regexp to match against the channels.
// 
// Returns generated access token string or an error if something went wrong..
func (c *Client) RequestSingleAccessToken(uid, pattern string) (token string, err error) {
	payload := []string{"AT", uid, pattern}
	token, err = c.performRequest(payload)
	return
}

type Worker struct {
	*socket
}

func NewWorker(uri string) (c *Worker, err error) {
	c = &Worker{}
	c.socket, err = newSocket("dlr", uri)
	return
}
