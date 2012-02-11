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
	"testing"
	"fmt"
)

var c *Client

type expectation struct {
	Name   string
	Action func() bool
	Expect func() bool
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
	var err error
	c, err = NewClient(fmt.Sprintf("wr://%s@127.0.0.1:8091/test", v.AccessToken()))
	if err != nil {
		t.Fatalf("Expected to connect the client, error: %v", err)
	}
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
