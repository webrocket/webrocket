package kosmonaut

import (
	"bytes"
	"github.com/webrocket/webrocket/pkg/webrocket"
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
	go ctx.NewWebsocketEndpoint(":8080").ListenAndServe()
	go ctx.NewBackendEndpoint(":8081").ListenAndServe()
	go ctx.NewAdminEndpoint(":8082").ListenAndServe()
	<-time.After(1e9)
}