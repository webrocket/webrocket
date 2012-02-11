package kosmonaut

import (
	"testing"
	"code.google.com/p/go.net/websocket"
	"fmt"
)

func TestWorkerFlow(t *testing.T) {
	w, err := NewWorker(fmt.Sprintf("wr://%s@127.0.0.1:8091/test", v.AccessToken()))
	if err != nil {
		t.Fatalf("Expected to connect the worker, error: %v", err)
	}
	go func() {
		ws, _ := websocket.Dial("ws://127.0.0.1:8090/test", "ws", "http://127.0.0.1/")
		token := v.GenerateSingleAccessToken("joe", ".*")
		var resp map[string]interface{}
		websocket.JSON.Receive(ws, &resp)
		websocket.JSON.Send(ws, map[string]interface{}{
			"auth": map[string]interface{}{
				"token": token,
			},
		})
		websocket.JSON.Receive(ws, &resp)
		websocket.JSON.Send(ws, map[string]interface{}{
			"trigger": map[string]interface{}{
				"event": "test",
				"data":  map[string]interface{}{"foo": "bar"},
			},
		})
	}()
	msg := <-w.Run()
	if msg.Event != "test" || msg.Data["foo"] != "bar" {
		t.Errorf("Expected to get the test event, got: %v", msg.Event)
	}
	w.Stop()
}

func TestWorkerWhenUnauthorized(t *testing.T) {
	w, _ := NewWorker("wr://foo@127.0.0.1:8091/test")
	for msg := range w.Run() {
		if msg.Error == nil || msg.Error.Error() != "402 Unauthorized" {
			t.Fatalf("Expected worker to be unauthorized")
		}
	}
}