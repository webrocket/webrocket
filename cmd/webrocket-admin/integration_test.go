package main

import (
	"bytes"
	webrocket "github.com/webrocket/webrocket/engine"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

var ctx *webrocket.Context

func init() {
	os.RemoveAll("./_testdata")
	ctx = webrocket.NewContext()
	ctx.SetLog(log.New(bytes.NewBuffer([]byte{}), "", log.LstdFlags))
	ctx.SetNodeName("test")
	ctx.SetStorageDir("./_testdata")
	ctx.Load()
	ctx.GenerateCookie(false)
	admin := ctx.NewAdminEndpoint(":8072")
	go admin.ListenAndServe()
	go ctx.NewWebsocketEndpoint(":8070").ListenAndServe()
	go ctx.NewBackendEndpoint(":8071").ListenAndServe()
	for !admin.IsAlive() {
		<-time.After(500 * time.Nanosecond)
	}
}

type cmdtest struct {
	args   []string
	expect *regexp.Regexp
}

var expectations = []cmdtest{
	{
		[]string{"list_vhosts"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"add_vhost", "foo"},
		regexp.MustCompile("invalid path"),
	}, {
		[]string{"add_vhost", "/hello"},
		regexp.MustCompile("/hello\n.{40}"),
	}, {
		[]string{"add_vhost", "/hello"},
		regexp.MustCompile("vhost already exists"),
	}, {
		[]string{"list_vhosts"},
		regexp.MustCompile("/hello"),
	}, {
		[]string{"show_vhost", "foo"},
		regexp.MustCompile("vhost doesn't exist"),
	}, {
		[]string{"show_vhost", "/hello"},
		regexp.MustCompile("/hello\n.{40}"),
	}, {
		[]string{"regenerate_vhost_token", "foo"},
		regexp.MustCompile("vhost doesn't exist"),
	}, {
		[]string{"regenerate_vhost_token", "/hello"},
		regexp.MustCompile(".{40}"),
	}, {
		[]string{"list_channels", "foo"},
		regexp.MustCompile("vhost doesn't exist"),
	}, {
		[]string{"list_channels", "/hello"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"add_channel", "/hello", "==="},
		regexp.MustCompile("invalid channel name"),
	}, {
		[]string{"add_channel", "foo", "world"},
		regexp.MustCompile("vhost doesn't exist"),
	}, {
		[]string{"add_channel", "/hello", "world"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"list_channels", "/hello"},
		regexp.MustCompile("world\t\\(0 subscribers\\)"),
	}, {
		[]string{"delete_channel", "/hello", "foo"},
		regexp.MustCompile("channel doesn't exist"),
	}, {
		[]string{"delete_channel", "foo", "world"},
		regexp.MustCompile("vhost doesn't exist"),
	}, {
		[]string{"delete_channel", "/hello", "world"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"add_channel", "/hello", "foo"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"add_channel", "/hello", "bar"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"clear_channels", "foo"},
		regexp.MustCompile("vhost doesn't exist"),
	}, {
		[]string{"clear_channels", "/hello"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"list_channels", "/hello"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"delete_vhost", "foo"},
		regexp.MustCompile("vhost doesn't exist"),
	}, {
		[]string{"delete_vhost", "/hello"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"list_vhosts"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"add_vhost", "/hello"},
		regexp.MustCompile("/hello\n.{40}"),
	}, {
		[]string{"add_vhost", "/world"},
		regexp.MustCompile("/world\n.{40}"),
	}, {
		[]string{"clear_vhosts"},
		regexp.MustCompile("^$"),
	}, {
		[]string{"list_vhosts"},
		regexp.MustCompile("^$"),
	},
}

func TestCommands(t *testing.T) {
	for _, exp := range expectations {
		var args []string
		args = append([]string{"-cookie=" + ctx.Cookie(), "-admin-addr=127.0.0.1:8072"}, exp.args...)
		cmd := exec.Command("../../webrocket-admin", args...)
		out, _ := cmd.CombinedOutput()
		if !exp.expect.Match(out) {
			argstr := strings.Join(exp.args, " ")
			t.Errorf("Expected `%s` to give correct output, got: %s", argstr, string(out))
		}
	}
}

// TODO: test list_workers command
