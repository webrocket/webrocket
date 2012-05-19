package kosmonaut

import (
	"bytes"
	webrocket "github.com/webrocket/webrocket/engine"
	"log"
	"os"
	"time"
)

var (
	v   *webrocket.Vhost
	ctx *webrocket.Context
)

func init() {
	os.RemoveAll("./_testdata")
	ctx = webrocket.NewContext()
	ctx.SetLog(log.New(bytes.NewBuffer([]byte{}), "", log.LstdFlags))
	ctx.SetNodeName("test")
	ctx.SetStorageDir("./_testdata")
	ctx.Load()
	v, _ = ctx.AddVhost("/test")
	ctx.GenerateCookie(false)
	backend := ctx.NewBackendEndpoint(":8091")
	go backend.ListenAndServe()
	go ctx.NewWebsocketEndpoint(":8090").ListenAndServe()
	go ctx.NewAdminEndpoint(":8092").ListenAndServe()
	for !backend.IsAlive() {
		<-time.After(500 * time.Nanosecond)
	}
}
