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
	"encoding/json"
	"errors"
	"github.com/bmizerany/pat"
	"net/http"
	"sync"
)

// adminHandler is a HTTP handler providing RESTful interface for
// the admin endpoint. 
type adminHandler struct {
	// Handler's multiplexer.
	mux *pat.PatternServeMux
}

var (
	// adminMtx is a mutex for admin context.
	adminMtx sync.Mutex
	// adminCtx is a pointer to current context.
	adminCtx *Context
	// adminMux is default admin multiplexer.
	adminMux *pat.PatternServeMux = pat.New()
)

// Admin handler's initializer
func init() {
	adminMux.Post("/:vhost/channels/:channel", http.HandlerFunc(adminAddChannel))
	adminMux.Get("/:vhost/channels/:channel", http.HandlerFunc(adminGetChannel))
	adminMux.Del("/:vhost/channels/:channel", http.HandlerFunc(adminDeleteChannel))
	adminMux.Del("/:vhost/channels", http.HandlerFunc(adminClearChannels))
	adminMux.Get("/:vhost/workers", http.HandlerFunc(adminListWorkers))
	adminMux.Get("/:vhost/channels", http.HandlerFunc(adminListChannels))
	adminMux.Put("/:vhost/token", http.HandlerFunc(adminRegenerateVhostToken))
	adminMux.Post("/:vhost", http.HandlerFunc(adminAddVhost))
	adminMux.Get("/:vhost", http.HandlerFunc(adminGetVhost))
	adminMux.Del("/:vhost", http.HandlerFunc(adminDeleteVhost))
	adminMux.Get("/", http.HandlerFunc(adminListVhosts))
	adminMux.Del("/", http.HandlerFunc(adminClearVhosts))
}

// Internal constructor
// -----------------------------------------------------------------------------

// newAdminHandler creates new handler for the specified context.
//
// ctx - Parent context.
//
// Returns created handler.
func newAdminHandler(ctx *Context) *adminHandler {
	adminMtx.Lock()
	defer adminMtx.Unlock()
	adminCtx = ctx
	return &adminHandler{mux: adminMux}
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
		adminCtx.log.Printf("admin: %s %s; %d %s", r.Method, r.RequestURI,
			code, http.StatusText(code))
	} else {
		adminCtx.log.Printf("admin: %s %s; %d %s: %s", r.Method, r.RequestURI,
			code, http.StatusText(code), err.Error())
	}
}

// authenticate checks if the request contains valid cookie.
//
// r - The request to be authenticated.
//
func (h *adminHandler) authenticate(r *http.Request) bool {
	cookie := r.Header.Get("X-WebRocket-Cookie")
	return cookie == adminCtx.Cookie()
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
	var code int

	if !h.authenticate(r) {
		code, _ = http.StatusForbidden, errors.New("access denied")
		w.WriteHeader(code)
		return
	}

	w.Header().Set("X-WebRocket-Cookie", adminCtx.Cookie())
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	h.mux.ServeHTTP(w, r)
	// TODO: log
}

// Admin interface actions
// -----------------------------------------------------------------------------

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
// GET /
//
func adminListVhosts(w http.ResponseWriter, r *http.Request) {
	data, i := make([]map[string]interface{}, len(adminCtx.vhosts)), 0
	for _, vhost := range adminCtx.Vhosts() {
		data[i] = map[string]interface{}{
			"path":        vhost.path,
			"accessToken": vhost.accessToken,
			"links": adminHypermediaLinks(
				[]string{"self", vhost.path},
			),
		}
		i += 1
	}
	w.WriteHeader(http.StatusOK)
	adminWriteData(w, "vhosts", data)
}

// adminAddVhost creates new vhost.
//
// POST /:vhost
//
func adminAddVhost(w http.ResponseWriter, r *http.Request) {
	path := "/" + r.URL.Query().Get(":vhost")
	if _, err := adminCtx.AddVhost(path); err != nil {
		adminWriteError(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Location", path)
	w.WriteHeader(http.StatusFound)
}

// adminGetVhost shows information about the vhost.
//
// GET /:vhost
//
func adminGetVhost(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	channels := map[string]interface{}{
		"size": len(vhost.Channels()),
	}
	data := map[string]interface{}{
		"path":        path,
		"accessToken": vhost.accessToken,
		"channels":    channels,
		"links": adminHypermediaLinks(
			[]string{"channels", path + "/channels"},
			[]string{"self", path},
		),
	}
	w.WriteHeader(http.StatusOK)
	adminWriteData(w, "vhost", data)
}

// adminDeleteVhost removes specified vhost.
//
// DELETE /:vhost
//
func adminDeleteVhost(w http.ResponseWriter, r *http.Request) {
	path := "/" + r.URL.Query().Get(":vhost")
	if err := adminCtx.DeleteVhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// adminClearVhosts removes all vhosts.
//
// DELETE /
//
func adminClearVhosts(w http.ResponseWriter, r *http.Request) {
	for path := range adminCtx.vhosts {
		adminCtx.DeleteVhost(path)
	}
	w.WriteHeader(http.StatusAccepted)
}

// adminRegenerateVhostToken generates new access token for the vhost.
//
// PUT /:vhost/token
//
func adminRegenerateVhostToken(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	vhost.GenerateAccessToken()
	w.Header().Set("Location", path)
	w.WriteHeader(http.StatusFound)
}

// adminListChannels shows list of channels from the specified vhost.
//
// GET /:vhost/channels
//
func adminListChannels(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	data, i := make([]map[string]interface{}, len(vhost.Channels())), 0
	for _, channel := range vhost.Channels() {
		data[i] = map[string]interface{}{
			"name": channel.name,
			"links": adminHypermediaLinks(
				[]string{"self", path + "/channels/" + channel.name},
				[]string{"vhost", path},
			),
		}
		i += 1
	}
	w.WriteHeader(http.StatusOK)
	adminWriteData(w, "channels", data)
}

// adminAddChannel creates new channel under the specified vhost.
//
// POST /:vhost/channels/:channel
//
func adminAddChannel(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	name := r.URL.Query().Get(":channel")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	kind := channelTypeFromName(name)
	if _, err = vhost.OpenChannel(name, kind); err != nil {
		adminWriteError(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Location", path+"/channels/"+name)
	w.WriteHeader(http.StatusFound)
	return
}

// adminGetChannel displays information about the channel.
//
// GET /:vhost/channels/:channel
//
func adminGetChannel(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var channel *Channel
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	name := r.URL.Query().Get(":channel")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	if channel, err = vhost.Channel(name); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	subscribers := map[string]interface{}{
		"size": len(channel.Subscribers()),
	}
	data := map[string]interface{}{
		"name":        channel.name,
		"subscribers": subscribers,
		"links": adminHypermediaLinks(
			[]string{"self", path + "/channels/" + name},
			[]string{"vhost", path},
			[]string{"subscribers", path + "/channels/" + name + "/subscribers"},
		),
	}
	w.WriteHeader(http.StatusOK)
	adminWriteData(w, "channel", data)
}

// adminDeleteChannels removes the channel.
//
// DELETE /:vhost/channels/:channel
//
func adminDeleteChannel(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	name := r.URL.Query().Get(":channel")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	if err = vhost.DeleteChannel(name); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// adminClearChannels removes all the channels.
//
// DELETE /:vhost/channels
//
func adminClearChannels(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	for name := range vhost.channels {
		vhost.DeleteChannel(name)
	}
	w.WriteHeader(http.StatusAccepted)
}

// adminListWorkers shows list of the active backend workers for the specified
// vhost.
//
// GET /:vhost/workers
//
func adminListWorkers(w http.ResponseWriter, r *http.Request) {
	var vhost *Vhost
	var err error
	path := "/" + r.URL.Query().Get(":vhost")
	if vhost, err = adminCtx.Vhost(path); err != nil {
		adminWriteError(w, http.StatusNotFound, err)
		return
	}
	data, i := make([]map[string]interface{}, len(vhost.lobby.Workers())), 0
	for _, worker := range vhost.lobby.Workers() {
		data[i] = map[string]interface{}{
			"id": worker.id,
			"links": adminHypermediaLinks(
				[]string{"self", path + "/workers/" + worker.id},
				[]string{"vhost", path},
			),
		}
		i += 1
	}
	w.WriteHeader(http.StatusOK)
	adminWriteData(w, "workers", data)
}

// adminHypermediaLinks generates map of links from the given list.
//
// links - list of links to pack
//
// Returns map of hypermedia links.
func adminHypermediaLinks(links ...[]string) (res []map[string]interface{}) {
	res = make([]map[string]interface{}, len(links))
	for i, link := range links {
		res[i] = map[string]interface{}{"rel": link[0], "href": link[1]}
	}
	return
}
