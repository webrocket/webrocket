package kosmonaut

import (
	"net"
	"encoding/json"
	"time"
)

// Timeout value for the client requests. 
const RequestTimeout = 5 * time.Second

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
	packet := pack(payload, c.Identity)
	deadline := time.Now().Add(RequestTimeout)
	conn.SetDeadline(deadline)
	conn.Write(packet)
	if response, err = recv(conn); err != nil {
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
	if len(frames) > 1 {
		switch frames[0] {
		case "OK":
			return "", nil
		case "ER": // Error
			return "", parseError(frames[1:])
		case "AT": // Single access token
			if len(frames) == 2 {
				token := frames[1]
				if len(token) == 128 {
					return token, nil
				}
			}
		}
	}
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