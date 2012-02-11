package kosmonaut

import (
	"testing"
	"fmt"
	"time"
)

func TestWorkerFlow(t *testing.T) {
	w, err := NewWorker(fmt.Sprintf("wr://%s@127.0.0.1:8081/test", v.AccessToken()))
	if err != nil {
		t.Fatalf("Expected to connect the worker, error: %v", err)
	}
	go func() {
		<-time.After(4 * time.Second)
		w.Stop()
		println("stopping")
	}()
	println("starting")
	for x := range w.Run() {
		fmt.Printf("%v\n", x)
	}
}

func TestWorkerWhenUnauthorized(t *testing.T) {
	w, _ := NewWorker("wr://foo@127.0.0.1:8081/test")
	defer func() {
		if r := recover(); r == nil || r != "402 Unauthorized" {
			t.Errorf("Expected to get unauthorized panic")
		}
	}
	w.Run()
	<-time.After(100 * time.Millisecond)
}