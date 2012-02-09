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
	"bytes"
	"fmt"
	"github.com/webrocket/webrocket/pkg/webrocket"
	"log"
	"os"
	"testing"
	"time"
)

var (
	c   *Client
	v   *webrocket.Vhost
	ctx *webrocket.Context
)

type expectation struct {
	Name   string
	Action func() bool
	Expect func() bool
}

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

func clientConnect(t *testing.T) {
	var err error
	c, err = NewClient(fmt.Sprintf("wr://%s@127.0.0.1:8081/test", v.AccessToken()))
	if err != nil {
		t.Fatalf("Expected to connect the client, error: %v", err)
	}
}

var expectations = []expectation{
	{
		"OpenChannel.1",
		func() bool {
			return c.OpenChannel("foo") == nil
		},
		func() bool {
			_, err := v.Channel("foo")
			return err == nil
		},
	}, {
		"OpenChannel.2",
		func() bool {
			err := c.OpenChannel("%%%")
			return err != nil && err.(*Error).Code == 451
		},
		func() bool {
			return true
		},
	}, {
		"Broadcast.1",
		func() bool {
			err := c.Broadcast("foobar", "test", map[string]interface{}{})
			return err != nil && err.(*Error).Code == 454
		},
		func() bool {
			return true
		},
	}, {
		"Broadcast.2",
		func() bool {
			err := c.Broadcast("foo", "test", map[string]interface{}{})
			return err == nil
		},
		func() bool {
			return true
		},
	}, {
		"CloseChannel.1",
		func() bool {
			return c.CloseChannel("foo") == nil
		},
		func() bool {
			_, err := v.Channel("foo")
			return err != nil
		},
	}, {
		"CloseChannel.2",
		func() bool {
			err := c.CloseChannel("foo")
			return err != nil && err.(*Error).Code == 454
		},
		func() bool {
			return true
		},
	}, {
		"RequestSingleAccessToken",
		func() bool {
			if token, err := c.RequestSingleAccessToken("joe", ".*"); err == nil {
				p, ok := v.ValidateSingleAccessToken(token)
				return ok && p.Uid() == "joe"
			}
			return false
		},
		func() bool {
			return true
		},
	},
}

func TestClientAPI(t *testing.T) {
	clientConnect(t)
	for _, expect := range expectations {
		if !expect.Action() {
			t.Errorf("Expected action `%s` to be performed properly", expect.Name)
			continue
		}
		if !expect.Expect() {
			t.Errorf("Expected test `%s` to pass", expect.Name)
		}
	}
}
