// Copyright (C) 2011 by Krzysztof Kowalik <chris@nu7hat.ch>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package engine

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"
)

// AdminEndpoint implements a wrapper for the http server instance which
// provides admin's RESTful interface. 
type AdminEndpoint struct {
	*http.Server

	// Context to which the endpoint belongs.
	ctx *Context
	// Information whether the endpoint is alive or not. 
	alive bool
	// Internal semaphore.
	mtx sync.Mutex
	// Internal logger.
	log *log.Logger
}

// Internal constructors
// -----------------------------------------------------------------------------

// newAdminEndpoint creates new admin endpoint configured to be bound to
// specified address. If no host specified in the address (eg. `:8080`),
// then will be bound to all available interfaces.
//
// ctx  - The parent context.
// addr - The host and port to which this endpoint will be bound.
//
// Returns new configured admin endpoint.
func newAdminEndpoint(ctx *Context, addr string) *AdminEndpoint {
	return &AdminEndpoint{
		ctx: ctx,
		log: ctx.log,
		Server: &http.Server{
			Addr:    addr,
			Handler: newAdminHandler(ctx),
		},
	}
}

// Exported
// -----------------------------------------------------------------------------

// Addr returns an address to which the endpoint is bound.
func (a *AdminEndpoint) Addr() string {
	return a.Server.Addr
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
//
// Returns an error if something went wrong.
func (a *AdminEndpoint) ListenAndServe() error {
	addr := a.Server.Addr
	if addr == "" {
		addr = ":http"
	}
	l, e := net.Listen("tcp", addr)
	if e != nil {
		return e
	}
	a.alive = true
	return a.Server.Serve(l)
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it expects
// HTTPS connections. Additionally, files containing a certificate and matching
// private key for the server must be provided. If the certificate is signed by
// a certificate authority, the certFile should be the concatenation of the
// server's certificate followed by the CA's certificate.

// One can use generate_cert.go in crypto/tls to generate cert.pem and key.pem.
//
// certFile - Path to the TLS certificate file.
// certKey  - Path to the certificate's private key.
//
// Returns an error if something went wrong.
func (a *AdminEndpoint) ListenAndServeTLS(certFile, certKey string) error {
	addr := a.Server.Addr
	if addr == "" {
		addr = ":https"
	}
	config := &tls.Config{
		Rand:       rand.Reader,
		NextProtos: []string{"http/1.1"},
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, certKey)
	if err != nil {
		return err
	}
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(conn, config)
	a.alive = true
	return a.Server.Serve(tlsListener)
}

// Returns true if this endpoint is activated.
func (a *AdminEndpoint) IsAlive() bool {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	return a.alive
}

// Kill stops execution of this endpoint.
func (a *AdminEndpoint) Kill() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.alive = false
}
