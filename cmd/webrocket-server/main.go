// This package implements executable for starting and preconfiguring
// single webrocket server node.
//
// Copyright (C) 2011 by Krzysztof Kowalik <chris@nu7hat.ch>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"flag"
	"fmt"
	stepper "github.com/nu7hatch/gostepper"
	"github.com/webrocket/webrocket/pkg/webrocket"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// Configuration variables.
var (
	// The backend endpoint bind address.
	BackendAddr string
	// The websocket endpoint bind address.
	WebsocketAddr string
	// The admin endpoint bind address.
	AdminAddr string
	// Custom node name.
	NodeName string
	// A path to the websocket endpoint certificate file.
	CertFile string
	// A path to the websocket endpoint key file.
	KeyFile string
	// A path to the The storage directory.
	StorageDir string
)

var (
	// The WebRocket main context.
	ctx *webrocket.Context
	// A stepper instance.
	s stepper.Stepper
)

func init() {
	flag.StringVar(&WebsocketAddr, "websocket-addr", ":8080", "websocket endpoint address")
	flag.StringVar(&BackendAddr, "backend-addr", ":8081", "backend endpoint address")
	flag.StringVar(&AdminAddr, "admin-addr", ":8082", "admin endpoint address")
	flag.StringVar(&NodeName, "node-name", "", "name of the node")
	flag.StringVar(&CertFile, "cert", "", "path to server certificate")
	flag.StringVar(&KeyFile, "key", "", "private key")
	flag.StringVar(&StorageDir, "storage-dir", "/var/lib/webrocket", "path to webrocket's internal data-store")
	flag.Parse()

	StorageDir, _ = filepath.Abs(StorageDir)
}

// SetupContext initializes global WebRocket context, loads configuration
// and all the vhosts data, and generates thean access cookie if its necessary.
func SetupContext() {
	s.Start("Initializing context")
	ctx = webrocket.NewContext()
	if err := ctx.SetStorageDir(StorageDir); err != nil {
		s.Fail(err.Error(), true)
	}
	if NodeName != "" {
		if err := ctx.SetNodeName(NodeName); err != nil {
			s.Fail(err.Error(), true)
		}
	}
	s.Ok()
	s.Start("Locking node")
	if err := ctx.Lock(); err != nil {
		s.Fail(err.Error(), true)
	}
	s.Ok()
	s.Start("Loading configuration")
	if err := ctx.Load(); err != nil {
		s.Fail(err.Error(), true)
	}
	s.Ok()
	s.Start("Generating cookie")
	if err := ctx.GenerateCookie(false); err != nil {
		s.Fail(err.Error(), true)
	}
	s.Ok()
}

// SetupEndpoint is a helper to configure and run a WebRocket endpoint.
//
// kind - The name of the endpoint.
// e    - The enpoint to be started.
//
func SetupEndpoint(kind string, e webrocket.Endpoint) {
	go func() {
		var err error
		s.Start("Starting %s", kind)
		if CertFile != "" && KeyFile != "" {
			err = e.ListenAndServeTLS(CertFile, KeyFile)
		} else {
			err = e.ListenAndServe()
		}
		if err != nil {
			s.Fail(err.Error(), true)
		}
	}()
	for !e.IsAlive() {
		<-time.After(500 * time.Nanosecond)
	}
	s.Ok()
}

// SignalTrap configures a handlers for various system signals, i.a.
// it stops the context and cleans everything up when the app is interrupted.
func SignalTrap() {
	var interrupted = make(chan os.Signal)
	signal.Notify(interrupted, syscall.SIGQUIT, syscall.SIGINT)
	<-interrupted
	fmt.Printf("\n\033[33mExiting...\033[0m\n")
	if ctx != nil {
		ctx.Kill()
	}
}

// DisplayAsciiArt as you can see it displays this amazing ASCII art
// spaceship drawing.
func DisplayAsciiArt() {
	fmt.Printf("\n")
	fmt.Printf(AsciiRocket)
	fmt.Printf("WebRocket v%s\n", webrocket.Version())
	fmt.Printf("Copyright (C) 2011-2012 by Krzysztof Kowalik and folks at Cubox.\n")
	fmt.Printf("Released under the AGPL. See http://webrocket.io/ for details.\n\n")
}

// DisplaySystemSettings shows all the information about the running
// webrocket node.
func DisplaySystemSettings() {
	fmt.Printf("\n")
	fmt.Printf("Node               : %s\n", ctx.NodeName())
	fmt.Printf("Cookie             : %s\n", ctx.Cookie())
	fmt.Printf("Data store dir     : %s\n", ctx.StorageDir())
	fmt.Printf("Websocket endpoint : ws://%s\n", WebsocketAddr)
	fmt.Printf("Backend endpoint   : wr://%s\n", BackendAddr)
	fmt.Printf("Admin endpoint     : http://%s\n", AdminAddr)

	fmt.Printf("\n\033[32mWebRocket has been launched!\033[0m\n")
}

func main() {
	DisplayAsciiArt()
	SetupContext()
	SetupEndpoint("backend endpoint", ctx.NewBackendEndpoint(BackendAddr))
	SetupEndpoint("websocket endpoint", ctx.NewWebsocketEndpoint(WebsocketAddr))
	SetupEndpoint("admin endpoint", ctx.NewAdminEndpoint(AdminAddr))
	DisplaySystemSettings()
	SignalTrap()
}

const AsciiRocket = `` +
	`            /\                                                                    ` + "\n" +
	`      ,    /  \      o               .        ___---___                    .      ` + "\n" +
	`          /    \            .              .--\        --.     .     .         .  ` + "\n" +
	`         /______\                        ./.;_.\     __/~ \.                      ` + "\n" +
	`   .    |        |                      /;  / '-'  __\    . \                     ` + "\n" +
	`        |        |    .        .       / ,--'     / .   .;   \        |           ` + "\n" +
	`        |________|                    | .|       /       __   |      -O-       .  ` + "\n" +
	`        |________|                   |__/    __ |  . ;   \ | . |      |           ` + "\n" +
	`       /|   ||   |\                  |      /  \\_    . ;| \___|                  ` + "\n" +
	`      / |   ||   | \    .    o       |      \  .~\\___,--'     |           .      ` + "\n" +
	`     /  |   ||   |  \                 |     | . ; ~~~~\_    __|                   ` + "\n" +
	`    /___|:::||:::|___\   |             \    \   .  .  ; \  /_/   .                ` + "\n" +
	`        |::::::::|      -O-        .    \   /         . |  ~/                  .  ` + "\n" +
	`         \::::::/         |    .          ~\ \   .      /  /~          o          ` + "\n" +
	`   o      ||__||       .                   ~--___ ; ___--~                        ` + "\n" +
	`            ||                        .          ---         .              .     ` + "\n" +
	`            ''                                                                    ` + "\n"
