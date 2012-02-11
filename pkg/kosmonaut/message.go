package kosmonaut

import (
	"errors"
	"encoding/json"
)

// Message represents single event incoming from the WebRocket server.
type Message struct {
	// The event name.
	Event string
	// Attached data.
	Data map[string]interface{}
	// An error if something went wrong.
	Error error
	// The backend worker via which the message has been received.
	worker *Worker
}

// parseMessage takes the raw message frames and extracts the message data
// from it. If message has invalid format then error message will be returned.
//
// rawmsg - The message frames to be parsed.
// worker - The worker which received this message.
//
// Returns parsed message instance.
func parseMessage(rawmsg []string, worker *Worker) (msg *Message) {
	var payload map[string]interface{}
	var ok bool
	var err error
	if len(rawmsg) != 1 {
		goto invalid
	}
	if err = json.Unmarshal([]byte(rawmsg[0]), &payload); err != nil {
		goto invalid
	}
	if len(payload) != 1 {
		goto invalid
	}
	msg = &Message{worker: worker}
	for event, data := range payload {
		msg.Event = event
		if msg.Data, ok = data.(map[string]interface{}); !ok {
			msg.Data = make(map[string]interface{})
		}
		break
	}
	return
invalid:
	msg = &Message{Error:errors.New("invalid message format")}
	return
}

// BroadcastReply broadcasts an event as a reply to the message. Reply will
// go to the specified channel.
//
// event   - An event name to be broadcasted.
// channel - The channel to broadcast to.
// data    - Data attached to the event.
//
// Returns an error if something went wrong.
func (msg *Message) BroadcastReply(event, channel string,
	data map[string]interface{}) (err error) {
	var c *Client
	if c, err = NewClient(msg.worker.URL.String()); err != nil {
		return
	}
	err = c.Broadcast(event, channel, data)
	return
}

// DirectReply sends a direct reply to the client who triggered the
// message.
//
// event - An event name to be triggered.
// data  - Data attached to the event.
//
// Returns an error if something went wrong.
func (msg *Message) DirectReply(event string, data map[string]interface{}) (
	err error) {
	return errors.New("not implemented")
}
