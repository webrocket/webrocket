package kosmonaut

import (
	"errors"
	"encoding/json"
)

type Message struct {
	Event  string
	Data   map[string]interface{}
	worker *Worker
}

func parseMessage(rawmsg []string, worker *Worker) (msg *Message, err error) {
	var payload map[string]interface{}
	var ok bool
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
	return nil, errors.New("invalid message payload")
}

func (msg *Message) BroadcastReply(event, channel string,
	data map[string]interface{}) (err error) {
	var c *Client
	if c, err = NewClient(msg.worker.URL.String()); err != nil {
		return
	}
	err = c.Broadcast(event, channel, data)
	return
}

func (msg *Message) DirectReply(event string, data map[string]interface{}) (
	err error) {
	return errors.New("not implemented")
}
