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

package webrocket

import (
	"encoding/json"
	"errors"
	"net/http"
)

// adminHandler is a HTTP handler providing RESTful interface for
// the admin endpoint. 
type adminHandler struct {
	// Parent context.
	ctx *Context
	// Handler's multiplexer.
	mux AdminServeMux
}

// Internal constructor
// -----------------------------------------------------------------------------

// newAdminHandler creates new handler for the specified context.
//
// ctx - Parent context.
//
// Returns created handler.
func newAdminHandler(ctx *Context) *adminHandler {
	return &adminHandler{ctx: ctx, mux: defaultAdminMux}
}

// Internal
// -----------------------------------------------------------------------------

// logStatus writes specified status information to the logs.
//
// r    - The request to be logged.
// code - Status code.
// err  - Encoundered error.
//
func (h *adminHandler) logStatus(r *http.Request, code int, err error) {
	if err == nil {
		h.ctx.log.Printf("admin: %s %s; %d %s", r.Method, r.RequestURI,
			code, http.StatusText(code))
	} else {
		h.ctx.log.Printf("admin: %s %s; %d %s: %s", r.Method, r.RequestURI,
			code, http.StatusText(code), err.Error())
	}
}

// authenticate checks if the request contains valid cookie.
//
// r - The request to be authenticated.
//
func (h *adminHandler) authenticate(r *http.Request) bool {
	cookie := r.Header.Get("X-WebRocket-Cookie")
	return cookie == h.ctx.Cookie()
}

// Exported
// -----------------------------------------------------------------------------

// ServeHTTP performs specified request and writes result to the response
// writer.
//
// w - The HTTP response writer.
// r - The request to be handled.
//
func (h *adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	var ok bool
	var fn adminHandlerFunc

	if !h.authenticate(r) {
		code, err = http.StatusForbidden, errors.New("access denied")
		w.WriteHeader(code)
		goto log
	}
	w.Header().Set("X-WebRocket-Cookie", h.ctx.Cookie())
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()
	if fn, ok = h.mux.Match(r.Method, r.RequestURI); !ok {
		code = http.StatusNotFound
		w.WriteHeader(code)
		goto log
	}
	code, err = fn(h.ctx, w, r)
log:
	h.logStatus(r, code, err)
}

// Admin interface actions
// -----------------------------------------------------------------------------

// List of default handlers provided by the admin interface. 
var defaultAdminMux AdminServeMux = map[string]adminHandlerFunc{
	"GET /vhosts":      adminListVhosts,
	"POST /vhosts":     adminAddVhost,
	"GET /vhost":       adminGetVhost,
	"DELETE /vhost":    adminDeleteVhost,
	"DELETE /vhosts":   adminClearVhosts,
	"PUT /vhost/token": adminRegenerateVhostToken,
	"GET /channels":    adminListChannels,
	"POST /channels":   adminAddChannel,
	"GET /channel":     adminGetChannel,
	"DELETE /channel":  adminDeleteChannel,
	"DELETE /channels": adminClearChannels,
	"GET /workers":     adminListWorkers,
}

// adminWriteError is a helper to writing error information to the response.
//
// w    - The response writer to write to.
// code - A HTTP status code
// err  - Encountered error
//
// Returns the same status code.
func adminWriteError(w http.ResponseWriter, code int, err error) int {
	data, _ := json.Marshal(map[string]string{"error": err.Error()})
	w.WriteHeader(code)
	w.Write(data)
	return code
}

// adminWriteData is a helper to write JSON-encoded interface to the
// response under specified namespace.
//
// w         - The response writer to write to.
// namespace - The namespace to write data under.
// x         - The aata to be serialized.
//
func adminWriteData(w http.ResponseWriter, namespace string, x interface{}) {
	data, _ := json.Marshal(map[string]interface{}{namespace: x})
	w.Write(data)
}

// adminListVhosts shows list of the vhosts.
//
// GET /vhosts
//
func adminListVhosts(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	data, i := make([]map[string]interface{}, len(ctx.vhosts)), 0
	for _, vhost := range ctx.Vhosts() {
		data[i] = map[string]interface{}{
			"self":        "/vhost?path=" + vhost.path,
			"path":        vhost.path,
			"accessToken": vhost.accessToken,
		}
		i += 1
	}
	code = http.StatusOK
	w.WriteHeader(code)
	adminWriteData(w, "vhosts", data)
	return
}

// adminAddVhost creates new vhost.
//
// POST /vhosts?path=[...]
//
func adminAddVhost(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	path := r.Form.Get("path")
	if _, err = ctx.AddVhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	code = http.StatusFound
	w.Header().Set("Location", "/vhost?path="+path)
	w.WriteHeader(code)
	return
}

// adminGetVhost shows information about the vhost.
//
// GET /vhost?path=[...]
//
func adminGetVhost(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	path := r.Form.Get("path")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	code = http.StatusOK
	channels := map[string]interface{}{
		"self": "/channels?vhost=" + path,
		"size": len(vhost.Channels()),
	}
	data := map[string]interface{}{
		"self":        "/vhost?path=" + path,
		"path":        path,
		"accessToken": vhost.accessToken,
		"channels":    channels,
	}
	w.WriteHeader(code)
	adminWriteData(w, "vhost", data)
	return
}

// adminDeleteVhost removes specified vhost.
//
// DELETE /vhost?path=[...]
//
func adminDeleteVhost(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	path := r.Form.Get("path")
	if err = ctx.DeleteVhost(path); err != nil {
		code = http.StatusNotFound
		adminWriteError(w, code, err)
		return
	}
	code = http.StatusAccepted
	w.WriteHeader(code)
	return
}

// adminClearVhosts removes all vhosts.
//
// DELETE /vhosts
//
func adminClearVhosts(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	for path := range ctx.vhosts {
		ctx.DeleteVhost(path)
	}
	code = http.StatusAccepted
	w.WriteHeader(code)
	return
}

// adminRegenerateVhostToken generates new access token for the vhost.
//
// PUT /vhost/token?vhost=[...]
//
func adminRegenerateVhostToken(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	path := r.Form.Get("path")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	vhost.GenerateAccessToken()
	code = http.StatusFound
	w.Header().Set("Location", "/vhost?path="+path)
	w.WriteHeader(code)
	return
}

// adminListChannels shows list of channels from the specified vhost.
//
// GET /channels?vhost=[...]
//
func adminListChannels(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	path := r.Form.Get("vhost")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	data, i := make([]map[string]interface{}, len(vhost.Channels())), 0
	for _, channel := range vhost.Channels() {
		data[i] = map[string]interface{}{
			"self":  "/channel?vhost=" + path,
			"vhost": "/vhost?path=" + path,
			"name":  channel.name,
		}
		i += 1
	}
	code = http.StatusOK
	w.WriteHeader(code)
	adminWriteData(w, "channels", data)
	return
}

// adminAddChannel creates new channel under the specified vhost.
//
// POST /channels?vhost=[...]&name=[...]
//
func adminAddChannel(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	path, name := r.Form.Get("vhost"), r.Form.Get("name")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	kind := channelTypeFromName(name)
	if _, err = vhost.OpenChannel(name, kind); err != nil {
		code = adminWriteError(w, http.StatusBadRequest, err)
		return
	}
	code = http.StatusFound
	w.Header().Set("Location", "/channel?vhost="+path+"&name="+name)
	w.WriteHeader(code)
	return
}

// adminGetChannel displays information about the channel.
//
// GET /channel?vhost=[...]&name=[...]
//
func adminGetChannel(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	var channel *Channel
	path, name := r.Form.Get("vhost"), r.Form.Get("name")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	if channel, err = vhost.Channel(name); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	subscribers := map[string]interface{}{
		"self": "/subscribers?vhost=" + path + "&channel=" + name,
		"size": len(channel.Subscribers()),
	}
	data := map[string]interface{}{
		"self":        "/channel?vhost=" + path + "&name=" + name,
		"vhost":       "/vhost?path=" + path,
		"name":        channel.name,
		"subscribers": subscribers,
	}
	code = http.StatusOK
	w.WriteHeader(code)
	adminWriteData(w, "channel", data)
	return
}

// adminDeleteChannels removes the channel.
//
// DELETE /channel?vhost=[...]&name=[...]
//
func adminDeleteChannel(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	path, name := r.Form.Get("vhost"), r.Form.Get("name")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	if err = vhost.DeleteChannel(name); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	code = http.StatusAccepted
	w.WriteHeader(code)
	return
}

// adminClearChannels removes all the channels.
//
// DELETE /channels?vhost=[...]
//
func adminClearChannels(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	path := r.Form.Get("vhost")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	for name := range vhost.channels {
		vhost.DeleteChannel(name)
	}
	code = http.StatusAccepted
	w.WriteHeader(code)
	return
}

// adminListWorkers shows list of the active backend workers for the specified
// vhost.
//
// GET /workers?vhost=[...]
//
func adminListWorkers(ctx *Context, w http.ResponseWriter, r *http.Request) (
	code int, err error) {
	var vhost *Vhost
	path := r.Form.Get("vhost")
	if vhost, err = ctx.Vhost(path); err != nil {
		code = adminWriteError(w, http.StatusNotFound, err)
		return
	}
	data, i := make([]map[string]interface{}, len(vhost.lobby.Workers())), 0
	for _, worker := range vhost.lobby.Workers() {
		data[i] = map[string]interface{}{
			"self":  "/worker?vhost=" + path + "&sid=" + worker.id,
			"vhost": "/vhost?path=" + path,
			"id":    worker.id,
		}
		i += 1
	}
	code = http.StatusOK
	w.WriteHeader(code)
	adminWriteData(w, "workers", data)
	return
}
